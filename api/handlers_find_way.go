package api

import (
	"encoding/json"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type findWayVertex struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type findWayEdge struct {
	ID     string `json:"id"`
	From   int    `json:"from"`
	To     int    `json:"to"`
	Weight int    `json:"weight"`
}

type findWayGraph struct {
	Vertices []findWayVertex `json:"vertices"`
	Edges    []findWayEdge   `json:"edges"`
}

type findWaySubmitRequest struct {
	SessionID       string   `json:"session_id"`
	RoundID         string   `json:"round_id"`
	SelectedEdgeIDs []string `json:"selected_edge_ids"`
	Timeout         bool     `json:"timeout"`
}

type findWayRoundMeta struct {
	Start  int `json:"start"`
	Finish int `json:"finish"`
}

type findWayStartResponse struct {
	SessionID   string       `json:"session_id"`
	RoundNumber int          `json:"round_number"`
	RoundID     string       `json:"round_id"`
	Graph       findWayGraph `json:"graph"`
	Start       int          `json:"start"`
	Finish      int          `json:"finish"`
}

type findWaySubmitResponse struct {
	IsValid            bool                  `json:"is_valid"`
	UserPathLength     int                   `json:"user_path_length"`
	ShortestPathLength int                   `json:"shortest_path_length"`
	ShortestPathEdges  []string              `json:"shortest_path_edges"`
	Message            string                `json:"message"`
	IsCorrect          bool                  `json:"is_correct"`
	NeedConfirmation   bool                  `json:"need_confirmation"`
	NextRound          *findWayStartResponse `json:"next_round,omitempty"`
	GameFinished       bool                  `json:"game_finished"`
	Statistics         *StatisticsResponse   `json:"statistics,omitempty"`
}

type findWayEndGameRequest struct {
	SessionID string `json:"session_id"`
}

type findWayConfirmRequest struct {
	SessionID string `json:"session_id"`
	Continue  bool   `json:"continue"`
}

func (m *Manager) hndlrFindWayStart(w http.ResponseWriter, r *http.Request) {
	session := m.sessionManager.CreateSession()
	firstRound := generateFindWayRound(m.sessionManager, session.SessionID)
	if firstRound == nil {
		m.sendErrorPage(w, http.StatusInternalServerError)
		return
	}

	m.send(w, firstRound)
}

func (m *Manager) hndlrFindWaySubmit(w http.ResponseWriter, r *http.Request) {
	var req findWaySubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	round, ok := m.sessionManager.GetRound(req.SessionID, req.RoundID)
	if !ok {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	graph, ok := round.SampleGraph.(findWayGraph)
	if !ok {
		m.sendErrorPage(w, http.StatusInternalServerError)
		return
	}
	roundMeta, ok := round.ChoiceGraphs.(findWayRoundMeta)
	if !ok {
		m.sendErrorPage(w, http.StatusInternalServerError)
		return
	}

	selected := req.SelectedEdgeIDs
	if req.Timeout {
		selected = []string{}
	}

	userLength, valid := calculateSelectedPathLength(graph, roundMeta.Start, roundMeta.Finish, selected)
	shortestLength, shortestEdges := shortestFindWayPath(graph, roundMeta.Start, roundMeta.Finish)
	success := valid && userLength == shortestLength

	message := "Путь не дошел до финиша."
	if success {
		message = "Отлично, это кратчайший путь!"
	} else if valid {
		message = "Путь построен, но можно короче."
	}

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

	var nextRound *findWayStartResponse
	if !needConfirmation && session.Status == "active" {
		nextRound = generateFindWayRound(m.sessionManager, req.SessionID)
	}

	m.send(w, findWaySubmitResponse{
		IsValid:            valid,
		UserPathLength:     userLength,
		ShortestPathLength: shortestLength,
		ShortestPathEdges:  shortestEdges,
		Message:            message,
		IsCorrect:          success,
		NeedConfirmation:   needConfirmation,
		NextRound:          nextRound,
		GameFinished:       false,
	})
}

func (m *Manager) hndlrFindWayConfirm(w http.ResponseWriter, r *http.Request) {
	var req findWayConfirmRequest
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
	nextRound := generateFindWayRound(m.sessionManager, req.SessionID)
	if nextRound == nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}
	m.send(w, map[string]any{
		"next_round": nextRound,
	})
}

func (m *Manager) hndlrFindWayEnd(w http.ResponseWriter, r *http.Request) {
	var req findWayEndGameRequest
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

func generateFindWayRound(sm *SessionManager, sessionID string) *findWayStartResponse {
	session, ok := sm.GetSession(sessionID)
	if !ok || session.Status != "active" {
		return nil
	}

	if session.Status != "active" {
        return nil
    }
	
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	graph := generateFindWayGraph(rnd)
	start := 0
	finish := len(graph.Vertices) - 1
	roundNumber := session.TotalRounds + 1
	roundID := uuid.New().String()

	round := &RoundData{
		RoundID:        roundID,
		RoundNumber:    roundNumber,
		CorrectIndices: nil,
		SampleGraph:    graph,
		ChoiceGraphs: findWayRoundMeta{
			Start:  start,
			Finish: finish,
		},
	}
	if !sm.SaveRound(sessionID, round) {
		return nil
	}

	return &findWayStartResponse{
		SessionID:   sessionID,
		RoundNumber: roundNumber,
		RoundID:     roundID,
		Graph:       graph,
		Start:       start,
		Finish:      finish,
	}
}

func generateFindWayGraph(rnd *rand.Rand) findWayGraph {
	vertexCount := rnd.Intn(4) + 7
	columns := (vertexCount + 1) / 2
	vertices := make([]findWayVertex, 0, vertexCount)
	top := make([]int, 0, columns)
	bottom := make([]int, 0, columns)
	xStep := 840
	if columns > 1 {
		xStep = 840 / (columns - 1)
	}
	for column := 0; len(vertices) < vertexCount; column++ {
		x := 80 + column*xStep + rnd.Intn(25) - 12
		top = append(top, len(vertices))
		vertices = append(vertices, findWayVertex{
			X: x,
			Y: 170 + rnd.Intn(35),
		})
		if len(vertices) >= vertexCount {
			break
		}
		bottom = append(bottom, len(vertices))
		vertices = append(vertices, findWayVertex{
			X: x + rnd.Intn(25) - 12,
			Y: 780 + rnd.Intn(35),
		})
	}

	edges := make([]findWayEdge, 0, vertexCount*2)
	addEdge := func(from, to int) {
		if from < 0 || to < 0 || from >= vertexCount || to >= vertexCount || from == to {
			return
		}
		for _, edge := range edges {
			if (edge.From == from && edge.To == to) || (edge.From == to && edge.To == from) {
				return
			}
		}
		edges = append(edges, findWayEdge{
			ID:     "e" + stringID(len(edges)),
			From:   from,
			To:     to,
			Weight: rnd.Intn(17) + 2,
		})
	}

	for i := 0; i+1 < len(top); i++ {
		addEdge(top[i], top[i+1])
	}
	for i := 0; i+1 < len(bottom); i++ {
		addEdge(bottom[i], bottom[i+1])
	}
	for i := 0; i < len(bottom); i++ {
		addEdge(top[i], bottom[i])
	}
	for i := 0; i+1 < len(bottom); i++ {
		if rnd.Intn(2) == 0 {
			addEdge(top[i], bottom[i+1])
		} else {
			addEdge(bottom[i], top[i+1])
		}
	}

	return findWayGraph{Vertices: vertices, Edges: edges}
}

func stringID(n int) string {
	if n == 0 {
		return "0"
	}
	digits := make([]byte, 0, 4)
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

func calculateSelectedPathLength(graph findWayGraph, start, finish int, selectedIDs []string) (int, bool) {
	edgeByID := map[string]findWayEdge{}
	for _, edge := range graph.Edges {
		edgeByID[edge.ID] = edge
	}

	current := start
	total := 0
	used := map[string]bool{}
	for _, id := range selectedIDs {
		edge, ok := edgeByID[id]
		if !ok || used[id] {
			return total, false
		}
		used[id] = true
		if edge.From == current {
			current = edge.To
		} else if edge.To == current {
			current = edge.From
		} else {
			return total, false
		}
		total += edge.Weight
	}

	return total, current == finish
}

func shortestFindWayPath(graph findWayGraph, start, finish int) (int, []string) {
	type adjacent struct {
		to     int
		weight int
		id     string
	}
	adj := make([][]adjacent, len(graph.Vertices))
	for _, edge := range graph.Edges {
		adj[edge.From] = append(adj[edge.From], adjacent{to: edge.To, weight: edge.Weight, id: edge.ID})
		adj[edge.To] = append(adj[edge.To], adjacent{to: edge.From, weight: edge.Weight, id: edge.ID})
	}

	dist := make([]int, len(graph.Vertices))
	prevVertex := make([]int, len(graph.Vertices))
	prevEdge := make([]string, len(graph.Vertices))
	visited := make([]bool, len(graph.Vertices))
	for i := range dist {
		dist[i] = math.MaxInt / 4
		prevVertex[i] = -1
	}
	dist[start] = 0

	for range graph.Vertices {
		v := -1
		for i := range graph.Vertices {
			if !visited[i] && (v == -1 || dist[i] < dist[v]) {
				v = i
			}
		}
		if v == -1 || v == finish {
			break
		}
		visited[v] = true
		for _, next := range adj[v] {
			if dist[v]+next.weight < dist[next.to] {
				dist[next.to] = dist[v] + next.weight
				prevVertex[next.to] = v
				prevEdge[next.to] = next.id
			}
		}
	}

	path := []string{}
	for v := finish; v != start && v >= 0; v = prevVertex[v] {
		if prevEdge[v] == "" {
			return 0, nil
		}
		path = append(path, prevEdge[v])
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return dist[finish], path
}
