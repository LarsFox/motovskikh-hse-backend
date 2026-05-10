package api

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type escapeEdge struct {
	ID   string `json:"id"`
	From int    `json:"from"`
	To   int    `json:"to"`
}

type escapeVertex struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type escapeGraph struct {
	Vertices []escapeVertex `json:"vertices"`
	Edges    []escapeEdge   `json:"edges"`
}

type escapeStartResponse struct {
	SessionID   string      `json:"session_id"`
	RoundNumber int         `json:"round_number"`
	RoundID     string      `json:"round_id"`
	Graph       escapeGraph `json:"graph"`
}

type escapeSubmitRequest struct {
	SessionID       string   `json:"session_id"`
	RoundID         string   `json:"round_id"`
	SelectedEdgeIDs []string `json:"selected_edge_ids"`
	Timeout         bool     `json:"timeout"`
}

type escapeSubmitResponse struct {
	IsValid          bool                 `json:"is_valid"`
	IsComplete       bool                 `json:"is_complete"`
	Message          string               `json:"message"`
	IsCorrect        bool                 `json:"is_correct"`
	NeedConfirmation bool                 `json:"need_confirmation"`
	NextRound        *escapeStartResponse `json:"next_round,omitempty"`
	GameFinished     bool                 `json:"game_finished"`
	Statistics       *StatisticsResponse  `json:"statistics,omitempty"`
}

type escapeConfirmRequest struct {
	SessionID string `json:"session_id"`
	Continue  bool   `json:"continue"`
}

type escapeEndGameRequest struct {
	SessionID string `json:"session_id"`
}



func (m *Manager) hndlrEscapeStart(w http.ResponseWriter, r *http.Request) {
	session := m.sessionManager.CreateSession()
	firstRound := generateEscapeRound(m.sessionManager, session.SessionID)
	if firstRound == nil {
		m.sendErrorPage(w, http.StatusInternalServerError)
		return
	}
	m.send(w, firstRound)
}

func (m *Manager) hndlrEscapeSubmit(w http.ResponseWriter, r *http.Request) {
	var req escapeSubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	round, ok := m.sessionManager.GetRound(req.SessionID, req.RoundID)
	if !ok {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	graph, ok := round.SampleGraph.(escapeGraph)
	if !ok {
		m.sendErrorPage(w, http.StatusInternalServerError)
		return
	}

	isValid, isComplete, message := validateEulerianPath(graph, req.SelectedEdgeIDs)

	// Если путь валидный, но неполный — не завершаем раунд
	if isValid && !isComplete {
		m.send(w, map[string]any{
			"is_valid":    true,
			"is_complete": false,
			"message":     message,
			"is_correct":  false,
		})
		return
	}

	if !isValid {
		m.send(w, map[string]any{
			"is_valid":    false,
			"is_complete": false,
			"message":     message,
			"is_correct":  false,
		})
		return
	}

	success := isValid && isComplete
	m.sessionManager.UpdateScore(req.SessionID, success)

	session, ok := m.sessionManager.GetSession(req.SessionID)
	if !ok {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	needConfirmation := session.TotalRounds > 0 && session.TotalRounds%15 == 0
	if needConfirmation {
		m.sessionManager.UpdateStatus(req.SessionID, "waiting_confirmation")
	}

	var nextRound *escapeStartResponse
	if !needConfirmation && session.Status == "active" {
		nextRound = generateEscapeRound(m.sessionManager, req.SessionID)
	}

	m.send(w, escapeSubmitResponse{
		IsValid:          isValid,
		IsComplete:       isComplete,
		Message:          message,
		IsCorrect:        success,
		NeedConfirmation: needConfirmation,
		NextRound:        nextRound,
		GameFinished:     false,
	})
}

func (m *Manager) hndlrEscapeConfirm(w http.ResponseWriter, r *http.Request) {
	var req escapeConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	if !req.Continue {
		totalRounds, correctAnswers, percentage, ok := m.sessionManager.GetStatistics(req.SessionID)
		if !ok {
			m.sendErrorPage(w, http.StatusBadRequest)
			return
		}
		m.sessionManager.UpdateStatus(req.SessionID, "finished")
		m.send(w, map[string]any{
			"statistics": map[string]any{
				"total_rounds":    totalRounds,
				"correct_answers": correctAnswers,
				"percentage":      percentage,
			},
		})
		return
	}

	m.sessionManager.UpdateStatus(req.SessionID, "active")
	nextRound := generateEscapeRound(m.sessionManager, req.SessionID)
	if nextRound == nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}
	m.send(w, map[string]any{
		"next_round": nextRound,
	})
}

func (m *Manager) hndlrEscapeEnd(w http.ResponseWriter, r *http.Request) {
	var req escapeEndGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	totalRounds, correctAnswers, percentage, ok := m.sessionManager.GetStatistics(req.SessionID)
	if !ok {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	m.sessionManager.UpdateStatus(req.SessionID, "finished")

	m.send(w, map[string]any{
		"total_rounds":    totalRounds,
		"correct_answers": correctAnswers,
		"percentage":      percentage,
	})
}





func generateEscapeRound(sm *SessionManager, sessionID string) *escapeStartResponse {
	session, ok := sm.GetSession(sessionID)
	if !ok || session.Status != "active" {
		return nil
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	var graph escapeGraph
	var graphHash string
	maxAttempts := 20

	for attempt := 0; attempt < maxAttempts; attempt++ {
		graph = generateEulerianGraph(rnd)

		if isSimpleCycle(graph) {
			continue
		}

		graphHash = hashGraph(graph)

		if isGraphUsed(sm, sessionID, graphHash) {
			continue
		}

		break
	}

	roundNumber := session.TotalRounds + 1
	roundID := uuid.New().String()

	round := &RoundData{
		RoundID:        roundID,
		RoundNumber:    roundNumber,
		CorrectIndices: nil,
		SampleGraph:    graph,
	}

	if !sm.SaveRound(sessionID, round) {
		return nil
	}

	saveGraphHash(sm, sessionID, graphHash)

	return &escapeStartResponse{
		SessionID:   sessionID,
		RoundNumber: roundNumber,
		RoundID:     roundID,
		Graph:       graph,
	}
}

func isSimpleCycle(graph escapeGraph) bool {
	if len(graph.Vertices) < 3 {
		return false
	}

	degrees := make([]int, len(graph.Vertices))
	for _, edge := range graph.Edges {
		degrees[edge.From]++
		degrees[edge.To]++
	}

	for _, deg := range degrees {
		if deg != 2 {
			return false
		}
	}

	return len(graph.Edges) == len(graph.Vertices)
}

func hashGraph(graph escapeGraph) string {
	edgeStrings := make([]string, 0, len(graph.Edges))
	for _, edge := range graph.Edges {
		if edge.From < edge.To {
			edgeStrings = append(edgeStrings, fmt.Sprintf("%d-%d", edge.From, edge.To))
		} else {
			edgeStrings = append(edgeStrings, fmt.Sprintf("%d-%d", edge.To, edge.From))
		}
	}
	sort.Strings(edgeStrings)
	hashInput := fmt.Sprintf("%d|%s", len(graph.Vertices), strings.Join(edgeStrings, ","))
	hash := md5.Sum([]byte(hashInput))
	return hex.EncodeToString(hash[:])
}

func isGraphUsed(sm *SessionManager, sessionID, graphHash string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists || session.GraphHistory == nil {
		return false
	}

	for _, h := range session.GraphHistory {
		if h == graphHash {
			return true
		}
	}
	return false
}

func saveGraphHash(sm *SessionManager, sessionID, graphHash string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return
	}
	if session.GraphHistory == nil {
		session.GraphHistory = make([]string, 0)
	}
	session.GraphHistory = append(session.GraphHistory, graphHash)
}

func generateEulerianGraph(rnd *rand.Rand) escapeGraph {
	vertexCount := rnd.Intn(7) + 4

	vertices := make([]escapeVertex, vertexCount)

	centerX, centerY := 400, 300
	radius := 200
	for i := 0; i < vertexCount; i++ {
		angle := (2 * math.Pi * float64(i)) / float64(vertexCount)
		offsetX := rnd.Intn(40) - 20
		offsetY := rnd.Intn(40) - 20
		vertices[i] = escapeVertex{
			X: centerX + int(float64(radius)*math.Cos(angle)) + offsetX,
			Y: centerY + int(float64(radius)*math.Sin(angle)) + offsetY,
		}
	}

	edges := make([]escapeEdge, 0)
	edgeID := 0
	degrees := make([]int, vertexCount)

	for i := 0; i < vertexCount; i++ {
		from := i
		to := (i + 1) % vertexCount
		edges = append(edges, escapeEdge{
			ID:   fmt.Sprintf("e%d", edgeID),
			From: from,
			To:   to,
		})
		edgeID++
		degrees[from]++
		degrees[to]++
	}

	extraEdgesCount := rnd.Intn(vertexCount) + 2

	for i := 0; i < extraEdgesCount; i++ {
		a := rnd.Intn(vertexCount)
		b := rnd.Intn(vertexCount)
		if a == b || edgeExistsInEscape(edges, a, b) {
			continue
		}

		oddBefore := 0
		for _, deg := range degrees {
			if deg%2 == 1 {
				oddBefore++
			}
		}

		oddAfter := oddBefore
		if (degrees[a]+1)%2 == 1 {
			oddAfter++
		} else {
			oddAfter--
		}
		if (degrees[b]+1)%2 == 1 {
			oddAfter++
		} else {
			oddAfter--
		}

		if oddAfter == 0 || oddAfter == 2 {
			edges = append(edges, escapeEdge{
				ID:   fmt.Sprintf("e%d", edgeID),
				From: a,
				To:   b,
			})
			edgeID++
			degrees[a]++
			degrees[b]++
		}
	}

	if len(edges) == vertexCount {
		for a := 0; a < vertexCount; a++ {
			for b := a + 1; b < vertexCount; b++ {
				if !edgeExistsInEscape(edges, a, b) && a != b {
					edges = append(edges, escapeEdge{
						ID:   fmt.Sprintf("e%d", edgeID),
						From: a,
						To:   b,
					})
					break
				}
			}
			if len(edges) > vertexCount {
				break
			}
		}
	}

	return escapeGraph{
		Vertices: vertices,
		Edges:    edges,
	}
}

func validateEulerianPath(graph escapeGraph, selectedEdgeIDs []string) (isValid bool, isComplete bool, message string) {
	if len(selectedEdgeIDs) == 0 {
		return true, false, "Выберите первое ребро"
	}

	edgeByID := make(map[string]escapeEdge)
	for _, edge := range graph.Edges {
		edgeByID[edge.ID] = edge
	}

	usedEdges := make(map[string]bool)
	currentVertex := -1

	for i, edgeID := range selectedEdgeIDs {
		edge, exists := edgeByID[edgeID]
		if !exists {
			return false, false, "Неизвестное ребро"
		}

		if usedEdges[edgeID] {
			return false, false, "Нельзя проходить по одному ребру дважды"
		}
		usedEdges[edgeID] = true

		if i == 0 {
			currentVertex = edge.To
		} else {
			if edge.From == currentVertex {
				currentVertex = edge.To
			} else if edge.To == currentVertex {
				currentVertex = edge.From
			} else {
				return false, false, "Путь прерван: ребро не соединено с текущей позицией"
			}
		}
	}

	allEdgesSelected := len(selectedEdgeIDs) == len(graph.Edges)

	if allEdgesSelected {
		return true, true, "Отлично! Вы нашли эйлеров путь!"
	}

	return true, false, fmt.Sprintf("Путь верный. Осталось выбрать %d рёбер", len(graph.Edges)-len(selectedEdgeIDs))
}

func edgeExistsInEscape(edges []escapeEdge, from, to int) bool {
	for _, e := range edges {
		if (e.From == from && e.To == to) || (e.From == to && e.To == from) {
			return true
		}
	}
	return false
}