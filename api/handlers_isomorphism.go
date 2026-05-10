package api

import (
    "encoding/json" 
    "math/rand"
    "net/http"
    "log"
    "fmt"
    "io"
    "bytes"
    
    "github.com/LarsFox/motovskikh-hse-backend/manager"
    "github.com/google/uuid"
)

//ответ на /start
type StartGameResponse struct {
    SessionID      string                   `json:"session_id"`
    RoundNumber    int                      `json:"round_number"`
    RoundID        string                   `json:"round_id"`
    TimeLimitSecs  int                      `json:"time_limit_secs"`
    SampleGraph    *manager.GraphData       `json:"sample_graph"`
    ChoiceGraphs   []manager.GraphData      `json:"choice_graphs"`
    CorrectIndices []int                    `json:"correct_indices"`
}

//запрос на проверку ответа
type SubmitAnswerRequest struct {
    SessionID      string `json:"session_id"`
    RoundID        string `json:"round_id"`
    SelectedIndices []int `json:"selectedIndices"`
    Timeout        bool   `json:"timeout"`
}

//ответ с результатом проверки
type SubmitAnswerResponse struct {
    IsCorrect        bool                       `json:"is_correct"`
    Feedback         map[string]string          `json:"feedback"` // "0": "correct"/"incorrect"/"missed"
    NextRound        *StartGameResponse         `json:"next_round,omitempty"`
    GameFinished     bool                       `json:"game_finished"`
    NeedConfirmation bool                       `json:"need_confirmation"`
    Statistics       *StatisticsResponse        `json:"statistics,omitempty"`
}

//статистика по игре
type StatisticsResponse struct {
    TotalRounds    int     `json:"total_rounds"`
    CorrectAnswers int     `json:"correct_answers"`
    Percentage     float64 `json:"percentage"`
}

//запрос на подтверждение продолжения
type ConfirmRequest struct {
    SessionID string `json:"session_id"`
    Continue  bool   `json:"continue"`
}

//ответ на подтверждение
type ConfirmResponse struct {
    NextRound  *StartGameResponse `json:"next_round,omitempty"`
    Statistics *StatisticsResponse `json:"statistics,omitempty"`
}

//обрабатывает подтверждение продолжения игры
func (m *Manager) hndlrConfirm(w http.ResponseWriter, r *http.Request) {
    var req ConfirmRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        m.sendErrorPage(w, http.StatusBadRequest)
        return
    }
    
    if !req.Continue {
        //не хочет продолжать — завершаем игру
        totalRounds, correctAnswers, percentage, _ := m.sessionManager.GetStatistics(req.SessionID)
        m.sessionManager.UpdateStatus(req.SessionID, "finished")
        
        m.send(w, ConfirmResponse{
            Statistics: &StatisticsResponse{
                TotalRounds:    totalRounds,
                CorrectAnswers: correctAnswers,
                Percentage:     percentage,
            },
        })
        return
    }
    
    //хочет продолжить — генерируем следующий раунд
    nextRound := generateNextRound(m.sessionManager, req.SessionID)
    m.send(w, ConfirmResponse{
        NextRound: nextRound,
    })
}

//один раунд игры
func (m *Manager) hndlrIsomorphismRound(w http.ResponseWriter, r *http.Request) {
    //генерим образец 5-8 вершин
    vertexCount := rand.Intn(4) + 5
    sampleGraph := manager.GenerateConnectedGraph(vertexCount)
    
    //генирим количество изоморфных графов 1-3
    isomorphicCount := rand.Intn(3) + 1
    
    //делаем 6 графов для выбор
    choiceGraphs := make([]manager.GraphData, 0, 6)
    correctIndices := make([]int, 0)
    
    // дообавляем изоморфные копии
    for i := 0; i < isomorphicCount; i++ {
        copyGraph := copyGraphWithNewCoordinates(sampleGraph)
        choiceGraphs = append(choiceGraphs, *copyGraph)
        correctIndices = append(correctIndices, i)
    }
    
    // дообавляем Неизоморфные графы
    for len(choiceGraphs) < 6 {
        nonIsomorphic := manager.GenerateConnectedGraph(vertexCount)
        choiceGraphs = append(choiceGraphs, *nonIsomorphic)
    }
    
    // перемешиваем графы между собой и обновляем индексы правильных ответов
    choiceGraphs, correctIndices = shuffleChoiceGraphs(choiceGraphs, correctIndices)
    

    m.send(w, map[string]any{
        "round_id":        uuid.New().String(),
        "sample_graph":    sampleGraph,
        "choice_graphs":   choiceGraphs,
        "correct_indices": correctIndices,
        "time_limit_secs": 60,
    })
}

//копирует структуру рёбер, генерирует новые координаты
func copyGraphWithNewCoordinates(original *manager.GraphData) *manager.GraphData {
    vertices := make([]manager.Vertex, original.VerticesCount)
    for i := 0; i < original.VerticesCount; i++ {
        vertices[i] = manager.Vertex{
            X: rand.Intn(1001),
            Y: rand.Intn(1001),
        }
    }
    
    return &manager.GraphData{
        VerticesCount: original.VerticesCount,
        EdgesCount:    original.EdgesCount,
        Vertices:      vertices,
        Edges:         original.Edges,
    }
}

// перемешивает графы и обновляет индексы правильных ответов
func shuffleChoiceGraphs(graphs []manager.GraphData, oldCorrectIndices []int) ([]manager.GraphData, []int) {
    // создаём пары оригинальный индекс, граф
    type pair struct {
        originalIdx int
        graph       manager.GraphData
    }
    
    pairs := make([]pair, len(graphs))
    for i, g := range graphs {
        pairs[i] = pair{originalIdx: i, graph: g}
    }
    
    // перемешиваем
    for i := len(pairs) - 1; i > 0; i-- {
        j := rand.Intn(i + 1)
        pairs[i], pairs[j] = pairs[j], pairs[i]
    }
    
    // собираем результат
    newGraphs := make([]manager.GraphData, len(graphs))
    newCorrectIndices := make([]int, 0)
    
    oldCorrectSet := make(map[int]bool)
    for _, idx := range oldCorrectIndices {
        oldCorrectSet[idx] = true
    }
    
    for newPos, p := range pairs {
        newGraphs[newPos] = p.graph
        if oldCorrectSet[p.originalIdx] {
            newCorrectIndices = append(newCorrectIndices, newPos)
        }
    }
    
    return newGraphs, newCorrectIndices
}




// начинает новую игру
func (m *Manager) hndlrStartGame(w http.ResponseWriter, r *http.Request) {
    // убрать ручные заголовки
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
    
    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusOK)
        return
    }

    // новая сеесия
    session := m.sessionManager.CreateSession()
    
    // первый раунд
    vertexCount := rand.Intn(4) + 5 // 5-8 вершин
    sampleGraph := manager.GenerateConnectedGraph(vertexCount)
    
    // количество изорфных
    isomorphicCount := rand.Intn(3) + 1
    
    // 6 на выбор
    choiceGraphs := make([]manager.GraphData, 0, 6)
    correctIndices := make([]int, 0)
    
    //изоморные копии
    for i := 0; i < isomorphicCount; i++ {
        copyGraph := copyGraphWithNewCoordinates(sampleGraph)
        choiceGraphs = append(choiceGraphs, *copyGraph)
        correctIndices = append(correctIndices, i)
    }
    
    //неизоморфные
    for len(choiceGraphs) < 6 {
        nonIsomorphic := manager.GenerateConnectedGraph(vertexCount)
        choiceGraphs = append(choiceGraphs, *nonIsomorphic)
    }
    
    // перемешиваем обновляем
    choiceGraphs, correctIndices = shuffleChoiceGraphs(choiceGraphs, correctIndices)
    
    //сохраняем раунд в сессию
    roundID := uuid.New().String()
    round := &RoundData{
        RoundID:        roundID,
        RoundNumber:    1,
        CorrectIndices: correctIndices,
        SampleGraph:    sampleGraph,
        ChoiceGraphs:   choiceGraphs,
    }
    m.sessionManager.SaveRound(session.SessionID, round)
    
    // отправляем ответ
    response := StartGameResponse{
        SessionID:      session.SessionID,
        RoundNumber:    1,
        RoundID:        roundID,
        TimeLimitSecs:  60,
        SampleGraph:    sampleGraph,
        ChoiceGraphs:   choiceGraphs,
        CorrectIndices: correctIndices,
    }
    
    m.send(w, response)
}



// проверяет ответ пользователя
func (m *Manager) hndlrSubmitAnswer(w http.ResponseWriter, r *http.Request) {
    fmt.Println("!!! hndlrSubmitAnswer was called !!!")  
    log.Println(">>> hndlrSubmitAnswer called")
    body, _ := io.ReadAll(r.Body)
    log.Printf("RAW BODY: %s", string(body))
    r.Body = io.NopCloser(bytes.NewBuffer(body))
    
    // Парсим JSON
    var req SubmitAnswerRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Println("JSON decode error:", err)
        m.sendErrorPage(w, http.StatusBadRequest)
        return
    }
    defer r.Body.Close()
    
    log.Printf("Request: session_id=%s, round_id=%s, selected=%v, timeout=%v", 
        req.SessionID, req.RoundID, req.SelectedIndices, req.Timeout)
    
    // получаем correct_indices из сессии
    correctIndices, ok := m.sessionManager.GetCorrectIndices(req.SessionID, req.RoundID)
    log.Printf("GetCorrectIndices: ok=%v, indices=%v", ok, correctIndices)
    
    if !ok {
        log.Println("Session or round not found")
        m.sendErrorPage(w, http.StatusBadRequest)
        return
    }
    
    // таймаут считаем неправильным ответ с пустыми выбранными
    selected := req.SelectedIndices
    if req.Timeout {
        selected = []int{}
    }
    
    // проверка
    selectedSet := make(map[int]bool)
    for _, idx := range selected {
        selectedSet[idx] = true
    }
    correctSet := make(map[int]bool)
    for _, idx := range correctIndices {
        correctSet[idx] = true
    
    }

    feedback := make(map[string]string)
    allCorrect := true
    hasIncorrect := false

    for i := 0; i < 6; i++ {
        isCorrect := correctSet[i]
        isSelected := selectedSet[i]
        
        if isCorrect && isSelected {
            feedback[string(rune(i+'0'))] = "correct"
        } else if !isCorrect && isSelected {
            feedback[string(rune(i+'0'))] = "incorrect"
            allCorrect = false
            hasIncorrect = true
        } else if isCorrect && !isSelected {
            feedback[string(rune(i+'0'))] = "missed"
            allCorrect = false
        }
    }

    log.Printf("Feedback generated: correctIndices=%v, selected=%v, feedback=%v", correctIndices, selected, feedback)
    
    
    // обновляем счет в сессии
    m.sessionManager.UpdateScore(req.SessionID, allCorrect && !hasIncorrect)
    
    // текущий номер раунда
    session, _ := m.sessionManager.GetSession(req.SessionID)
    currentRoundNumber := session.TotalRounds
    // нужен ли запрос подтверждения после 15, 30, 45... раундов?
    needConfirmation := currentRoundNumber%15 == 0 && currentRoundNumber > 0
    log.Printf("Round number: %d, needConfirmation: %v", currentRoundNumber, needConfirmation)
    
    // генерим следующий раунд если не нужна пауза на подтверждение
    var nextRound *StartGameResponse
    
    if !needConfirmation && session.Status == "active" {
        nextRound = generateNextRound(m.sessionManager, req.SessionID)
    }
    

    response := SubmitAnswerResponse{
        IsCorrect:        allCorrect && !hasIncorrect,
        Feedback:         feedback,
        NeedConfirmation: needConfirmation,
        GameFinished:     false,
    }
    
    if nextRound != nil {
        response.NextRound = nextRound
    }
    
    m.send(w, response)
}

//генерирует следующий раунд для сессии
func generateNextRound(sm *SessionManager, sessionID string) *StartGameResponse {
    session, ok := sm.GetSession(sessionID)
    if !ok || session.Status != "active" {
        return nil
    }
    
    // новый раунд аналогично /start
    vertexCount := rand.Intn(4) + 5
    sampleGraph := manager.GenerateConnectedGraph(vertexCount)
    
    isomorphicCount := rand.Intn(3) + 1
    
    choiceGraphs := make([]manager.GraphData, 0, 6)
    correctIndices := make([]int, 0)
    
    for i := 0; i < isomorphicCount; i++ {
        copyGraph := copyGraphWithNewCoordinates(sampleGraph)
        choiceGraphs = append(choiceGraphs, *copyGraph)
        correctIndices = append(correctIndices, i)
    }
    
    for len(choiceGraphs) < 6 {
        nonIsomorphic := manager.GenerateConnectedGraph(vertexCount)
        choiceGraphs = append(choiceGraphs, *nonIsomorphic)
    }
    
    choiceGraphs, correctIndices = shuffleChoiceGraphs(choiceGraphs, correctIndices)
    
    roundID := uuid.New().String()
    newRoundNumber := session.TotalRounds + 1
    
    round := &RoundData{
        RoundID:        roundID,
        RoundNumber:    newRoundNumber,
        CorrectIndices: correctIndices,
        SampleGraph:    sampleGraph,
        ChoiceGraphs:   choiceGraphs,
    }
    
    sm.SaveRound(sessionID, round)
    log.Printf("New round number: %d", newRoundNumber)
    
    return &StartGameResponse{
        SessionID:      sessionID,
        RoundNumber:    newRoundNumber,
        RoundID:        roundID,
        TimeLimitSecs:  60,
        SampleGraph:    sampleGraph,
        ChoiceGraphs:   choiceGraphs,
        CorrectIndices: correctIndices,
    }
}

// !!!!!!только для отладки!!!!!!!!
func (m *Manager) hndlrDebugSessions(w http.ResponseWriter, r *http.Request) {
    m.sessionManager.mu.RLock()
    defer m.sessionManager.mu.RUnlock()
    
    sessions := make([]string, 0, len(m.sessionManager.sessions))
    for id := range m.sessionManager.sessions {
        sessions = append(sessions, id)
    }
    
    m.send(w, map[string]interface{}{
        "sessions": sessions,
        "count":    len(sessions),
    })
}


// запрос на завершение игры
type EndGameRequest struct {
    SessionID string `json:"session_id"`
}

// ответ со статистикой
type EndGameResponse struct {
    TotalRounds    int     `json:"total_rounds"`
    CorrectAnswers int     `json:"correct_answers"`
    Percentage     float64 `json:"percentage"`
}

//завершает игру и возвращает статистику
func (m *Manager) hndlrEndGame(w http.ResponseWriter, r *http.Request) {
    var req EndGameRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        m.sendErrorPage(w, http.StatusBadRequest)
        return
    }
    
    totalRounds, correctAnswers, percentage, ok := m.sessionManager.GetStatistics(req.SessionID)
    if !ok {
        m.sendErrorPage(w, http.StatusBadRequest)
        return
    }
    
    // Помечаем сессию как завершённую
    m.sessionManager.UpdateStatus(req.SessionID, "finished")
    
    response := EndGameResponse{
        TotalRounds:    totalRounds,
        CorrectAnswers: correctAnswers,
        Percentage:     percentage,
    }
    
    m.send(w, response)
}