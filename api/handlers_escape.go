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

const escapeMaxRounds = 50

type escapeStartResponse struct {
	SessionID   string      `json:"session_id"`
	RoundNumber int         `json:"round_number"`
	MaxRounds   int         `json:"max_rounds"`
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

type escapeDifficulty struct {
	MinVertices      int
	MaxVertices      int
	MotifComplexity  int
	ExtraEdgeMin     int
	ExtraEdgeMax     int
	CrossLinkMin     int
	CrossLinkMax     int
	ClusterEdgeChance int
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

	gameFinished := false
	var statistics *StatisticsResponse

	if success && round.RoundNumber >= escapeMaxRounds {
		m.sessionManager.UpdateStatus(req.SessionID, "finished")
		totalRounds, correctAnswers, percentage, statOk := m.sessionManager.GetStatistics(req.SessionID)
		if statOk {
			statistics = &StatisticsResponse{
				TotalRounds:    totalRounds,
				CorrectAnswers: correctAnswers,
				Percentage:     percentage,
			}
		}
		gameFinished = true
	}

	needConfirmation := !gameFinished && session.TotalRounds > 0 && session.TotalRounds%15 == 0
	if needConfirmation {
		m.sessionManager.UpdateStatus(req.SessionID, "waiting_confirmation")
	}

	var nextRound *escapeStartResponse
	if !gameFinished && !needConfirmation && session.Status == "active" {
		nextRound = generateEscapeRound(m.sessionManager, req.SessionID)
	}

	m.send(w, escapeSubmitResponse{
		IsValid:          isValid,
		IsComplete:       isComplete,
		Message:          message,
		IsCorrect:        success,
		NeedConfirmation: needConfirmation,
		NextRound:        nextRound,
		GameFinished:     gameFinished,
		Statistics:       statistics,
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
	session, ok := m.sessionManager.GetSession(req.SessionID)
	if !ok {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}
	if session.TotalRounds >= escapeMaxRounds {
		totalRounds, correctAnswers, percentage, statOk := m.sessionManager.GetStatistics(req.SessionID)
		if !statOk {
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

	if session.TotalRounds >= escapeMaxRounds {
		return nil
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	roundNumber := session.TotalRounds + 1
	if roundNumber > escapeMaxRounds {
		return nil
	}

	var graph escapeGraph
	var graphHash string
	maxAttempts := 20

	for attempt := 0; attempt < maxAttempts; attempt++ {
		graph = generateEulerianGraph(rnd, roundNumber)

		if isSimpleCycle(graph) {
			continue
		}

		graphHash = hashGraph(graph)

		if isGraphUsed(sm, sessionID, graphHash) {
			continue
		}

		break
	}

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
		MaxRounds:   escapeMaxRounds,
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

func escapeDifficultyForRound(roundNumber int) escapeDifficulty {
	if roundNumber < 1 {
		roundNumber = 1
	}
	band := (roundNumber - 1) / 5
	if band > 9 {
		band = 9
	}
	tiers := []escapeDifficulty{
		{MinVertices: 6, MaxVertices: 8, MotifComplexity: 1, ExtraEdgeMin: 0, ExtraEdgeMax: 2, CrossLinkMin: 1, CrossLinkMax: 1, ClusterEdgeChance: 32},
		{MinVertices: 7, MaxVertices: 9, MotifComplexity: 2, ExtraEdgeMin: 1, ExtraEdgeMax: 3, CrossLinkMin: 1, CrossLinkMax: 2, ClusterEdgeChance: 36},
		{MinVertices: 8, MaxVertices: 10, MotifComplexity: 2, ExtraEdgeMin: 2, ExtraEdgeMax: 4, CrossLinkMin: 2, CrossLinkMax: 3, ClusterEdgeChance: 40},
		{MinVertices: 9, MaxVertices: 12, MotifComplexity: 3, ExtraEdgeMin: 2, ExtraEdgeMax: 5, CrossLinkMin: 2, CrossLinkMax: 3, ClusterEdgeChance: 44},
		{MinVertices: 10, MaxVertices: 13, MotifComplexity: 3, ExtraEdgeMin: 3, ExtraEdgeMax: 6, CrossLinkMin: 2, CrossLinkMax: 4, ClusterEdgeChance: 48},
		{MinVertices: 11, MaxVertices: 15, MotifComplexity: 4, ExtraEdgeMin: 4, ExtraEdgeMax: 7, CrossLinkMin: 3, CrossLinkMax: 4, ClusterEdgeChance: 52},
		{MinVertices: 12, MaxVertices: 16, MotifComplexity: 4, ExtraEdgeMin: 4, ExtraEdgeMax: 8, CrossLinkMin: 3, CrossLinkMax: 5, ClusterEdgeChance: 56},
		{MinVertices: 13, MaxVertices: 18, MotifComplexity: 5, ExtraEdgeMin: 5, ExtraEdgeMax: 9, CrossLinkMin: 3, CrossLinkMax: 5, ClusterEdgeChance: 60},
		{MinVertices: 14, MaxVertices: 20, MotifComplexity: 5, ExtraEdgeMin: 6, ExtraEdgeMax: 10, CrossLinkMin: 4, CrossLinkMax: 6, ClusterEdgeChance: 62},
		{MinVertices: 15, MaxVertices: 22, MotifComplexity: 6, ExtraEdgeMin: 7, ExtraEdgeMax: 12, CrossLinkMin: 4, CrossLinkMax: 6, ClusterEdgeChance: 65},
	}
	return tiers[band]
}

func generateEulerianGraph(rnd *rand.Rand, roundNumber int) escapeGraph {
	const maxBuildAttempts = 25
	diff := escapeDifficultyForRound(roundNumber)

	for attempt := 0; attempt < maxBuildAttempts; attempt++ {
		vertexRange := diff.MaxVertices - diff.MinVertices + 1
		vertexCount := diff.MinVertices + rnd.Intn(vertexRange)
		vertices := makeEscapeVertices(rnd, vertexCount)
		edges := make([]escapeEdge, 0, vertexCount*2)
		degrees := make([]int, vertexCount)
		edgeID := 0

		// Build a connected backbone (random spanning tree).
		for v := 1; v < vertexCount; v++ {
			parent := rnd.Intn(v)
			edges, edgeID = appendEscapeEdge(edges, edgeID, degrees, parent, v)
		}

		// Add motif edges so topology is not a simple "ring with chords".
		motif := rnd.Intn(3)
		switch motif {
		case 0:
			// Figure-eight feel: two local cycles touching by random connectors.
			repeats := rnd.Intn(diff.MotifComplexity) + 2
			for i := 0; i < repeats; i++ {
				a := rnd.Intn(vertexCount)
				b := rnd.Intn(vertexCount)
				c := rnd.Intn(vertexCount)
				if a != b && b != c && a != c {
					edges, edgeID = appendEscapeEdgeIfMissing(edges, edgeID, degrees, a, b)
					edges, edgeID = appendEscapeEdgeIfMissing(edges, edgeID, degrees, b, c)
					edges, edgeID = appendEscapeEdgeIfMissing(edges, edgeID, degrees, a, c)
				}
			}
		case 1:
			// Ladder/maze feel: short links and occasional skips.
			for i := 0; i < vertexCount-1; i++ {
				edges, edgeID = appendEscapeEdgeIfMissing(edges, edgeID, degrees, i, i+1)
				skipChance := 30 + diff.MotifComplexity*10
				if i+2 < vertexCount && rnd.Intn(100) < skipChance {
					edges, edgeID = appendEscapeEdgeIfMissing(edges, edgeID, degrees, i, i+2)
				}
			}
		default:
			// Dumbbell-ish feel: denser halves plus sparse cross links.
			mid := vertexCount / 2
			for i := 0; i < mid; i++ {
				for j := i + 1; j < mid; j++ {
					if rnd.Intn(100) < diff.ClusterEdgeChance {
						edges, edgeID = appendEscapeEdgeIfMissing(edges, edgeID, degrees, i, j)
					}
				}
			}
			for i := mid; i < vertexCount; i++ {
				for j := i + 1; j < vertexCount; j++ {
					if rnd.Intn(100) < diff.ClusterEdgeChance {
						edges, edgeID = appendEscapeEdgeIfMissing(edges, edgeID, degrees, i, j)
					}
				}
			}
			crossLinks := diff.CrossLinkMin + rnd.Intn(diff.CrossLinkMax-diff.CrossLinkMin+1)
			for i := 0; i < crossLinks; i++ {
				a := rnd.Intn(mid)
				b := mid + rnd.Intn(vertexCount-mid)
				edges, edgeID = appendEscapeEdgeIfMissing(edges, edgeID, degrees, a, b)
			}
		}

		// Additional randomization for variety.
		extra := diff.ExtraEdgeMin + rnd.Intn(diff.ExtraEdgeMax-diff.ExtraEdgeMin+1)
		for i := 0; i < extra; i++ {
			a := rnd.Intn(vertexCount)
			b := rnd.Intn(vertexCount)
			edges, edgeID = appendEscapeEdgeIfMissing(edges, edgeID, degrees, a, b)
		}

		// Enforce Euler condition: exactly 0 or 2 odd vertices.
		if !fixEulerParity(rnd, &edges, &edgeID, degrees) {
			continue
		}

		graph := escapeGraph{
			Vertices: vertices,
			Edges:    edges,
		}

		if isSimpleCycle(graph) {
			continue
		}

		return graph
	}

	// Fallback (very unlikely): return a small guaranteed Eulerian chain-like graph.
	fallbackVertices := makeEscapeVertices(rnd, 6)
	fallbackEdges := make([]escapeEdge, 0, 7)
	fallbackDegrees := make([]int, 6)
	edgeID := 0
	for i := 0; i < 5; i++ {
		fallbackEdges, edgeID = appendEscapeEdge(fallbackEdges, edgeID, fallbackDegrees, i, i+1)
	}
	fallbackEdges, edgeID = appendEscapeEdgeIfMissing(fallbackEdges, edgeID, fallbackDegrees, 1, 3)
	fallbackEdges, _ = appendEscapeEdgeIfMissing(fallbackEdges, edgeID, fallbackDegrees, 2, 4)

	return escapeGraph{
		Vertices: fallbackVertices,
		Edges:    fallbackEdges,
	}
}

func makeEscapeVertices(rnd *rand.Rand, vertexCount int) []escapeVertex {
	vertices := make([]escapeVertex, vertexCount)
	cols := int(math.Ceil(math.Sqrt(float64(vertexCount))))
	if cols < 2 {
		cols = 2
	}
	cellW, cellH := 700/cols, 460/cols
	startX, startY := 70, 70

	for i := 0; i < vertexCount; i++ {
		row := i / cols
		col := i % cols
		jitterX := rnd.Intn(50) - 25
		jitterY := rnd.Intn(50) - 25
		vertices[i] = escapeVertex{
			X: startX + col*cellW + jitterX,
			Y: startY + row*cellH + jitterY,
		}
	}

	return vertices
}

func appendEscapeEdge(edges []escapeEdge, edgeID int, degrees []int, from, to int) ([]escapeEdge, int) {
	edges = append(edges, escapeEdge{
		ID:   fmt.Sprintf("e%d", edgeID),
		From: from,
		To:   to,
	})
	degrees[from]++
	degrees[to]++

	return edges, edgeID + 1
}

func appendEscapeEdgeIfMissing(edges []escapeEdge, edgeID int, degrees []int, from, to int) ([]escapeEdge, int) {
	if from == to || edgeExistsInEscape(edges, from, to) {
		return edges, edgeID
	}

	return appendEscapeEdge(edges, edgeID, degrees, from, to)
}

func oddEscapeVertices(degrees []int) []int {
	odd := make([]int, 0)
	for v, deg := range degrees {
		if deg%2 == 1 {
			odd = append(odd, v)
		}
	}
	return odd
}

func fixEulerParity(rnd *rand.Rand, edges *[]escapeEdge, edgeID *int, degrees []int) bool {
	const maxAttempts = 80

	for attempts := 0; attempts < maxAttempts; attempts++ {
		odd := oddEscapeVertices(degrees)
		if len(odd) == 0 || len(odd) == 2 {
			return true
		}

		if len(odd) < 2 {
			return false
		}

		// Shuffle odd vertices for less predictable pairing.
		rnd.Shuffle(len(odd), func(i, j int) {
			odd[i], odd[j] = odd[j], odd[i]
		})

		paired := false
		for i := 0; i < len(odd) && !paired; i++ {
			for j := i + 1; j < len(odd); j++ {
				a, b := odd[i], odd[j]
				if a == b || edgeExistsInEscape(*edges, a, b) {
					continue
				}
				*edges, *edgeID = appendEscapeEdge(*edges, *edgeID, degrees, a, b)
				paired = true
				break
			}
		}

		// If odd vertices are fully interconnected, try a random non-existing edge.
		if !paired {
			vertexCount := len(degrees)
			placed := false
			for tries := 0; tries < vertexCount*2; tries++ {
				a := rnd.Intn(vertexCount)
				b := rnd.Intn(vertexCount)
				if a == b || edgeExistsInEscape(*edges, a, b) {
					continue
				}
				*edges, *edgeID = appendEscapeEdge(*edges, *edgeID, degrees, a, b)
				placed = true
				break
			}
			if !placed {
				return false
			}
		}
	}

	return false
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
	currentVertices := make(map[int]struct{})

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
			// First edge can be traversed from either endpoint.
			currentVertices[edge.From] = struct{}{}
			currentVertices[edge.To] = struct{}{}
		} else {
			nextVertices := make(map[int]struct{})
			if _, ok := currentVertices[edge.From]; ok {
				nextVertices[edge.To] = struct{}{}
			}
			if _, ok := currentVertices[edge.To]; ok {
				nextVertices[edge.From] = struct{}{}
			}
			if len(nextVertices) == 0 {
				return false, false, "Путь прерван: ребро не соединено с текущей позицией"
			}
			currentVertices = nextVertices
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