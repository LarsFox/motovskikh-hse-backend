package api

import (
    "sync"
    "time"
	"log" 
    
    "github.com/google/uuid"
)

// хранит информацию об одном раунде
type RoundData struct {
    RoundID        string
    RoundNumber    int
    CorrectIndices []int
    SampleGraph    interface{} // *manager.GraphData
    ChoiceGraphs   interface{} // []manager.GraphData
}

//хранит состояние игровой сессии
type GameSession struct {
    SessionID      string
    Rounds         map[string]*RoundData // key=round_id
    CurrentRoundID string
    TotalRounds    int
    CorrectAnswers int
    Status         string // "active", "waiting_confirmation", "finished"
    CreatedAt      time.Time
}

//управляет сессиями в памяти
type SessionManager struct {
    sessions map[string]*GameSession
    mu       sync.RWMutex
}

//создаёт новый менеджер сессий
func NewSessionManager() *SessionManager {
    return &SessionManager{
        sessions: make(map[string]*GameSession),
    }
}

//создаёт новую сессию
func (sm *SessionManager) CreateSession() *GameSession {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    sessionID := uuid.New().String()
    session := &GameSession{
        SessionID:      sessionID,
        Rounds:         make(map[string]*RoundData),
        TotalRounds:    0,
        CorrectAnswers: 0,
        Status:         "active",
        CreatedAt:      time.Now(),
    }
    
    sm.sessions[sessionID] = session
    return session
}

//возвращает сессию по ID
func (sm *SessionManager) GetSession(sessionID string) (*GameSession, bool) {
    sm.mu.RLock()
    defer sm.mu.RUnlock()
    
    session, exists := sm.sessions[sessionID]
    return session, exists
}

//сохраняет раунд в сессии
func (sm *SessionManager) SaveRound(sessionID string, round *RoundData) bool {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    

    log.Printf("SaveRound: session=%s, roundID=%s, roundNumber=%d", sessionID, round.RoundID, round.RoundNumber)
    
    session, exists := sm.sessions[sessionID]
    if !exists {
        return false
    }
    
    session.Rounds[round.RoundID] = round
    session.CurrentRoundID = round.RoundID
    session.TotalRounds = round.RoundNumber


    log.Printf("Round saved. Total rounds in session now: %d", len(session.Rounds))
    
    return true
}

//возвращает correct_indices по round_id
func (sm *SessionManager) GetCorrectIndices(sessionID, roundID string) ([]int, bool) {
    sm.mu.RLock()
    defer sm.mu.RUnlock()
    
    session, exists := sm.sessions[sessionID]
    if !exists {
        return nil, false
    }
    
    round, exists := session.Rounds[roundID]
    if !exists {
        return nil, false
    }
    
    return round.CorrectIndices, true
}

func getKeys(m map[string]*RoundData) []string {
    keys := make([]string, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    return keys
}

//обновляет счёт в сессии
func (sm *SessionManager) UpdateScore(sessionID string, isCorrect bool) bool {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    session, exists := sm.sessions[sessionID]
    if !exists {
        return false
    }
    
    if isCorrect {
        session.CorrectAnswers++
    }
    return true
}

//обновляет статус сессии
func (sm *SessionManager) UpdateStatus(sessionID, status string) bool {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    session, exists := sm.sessions[sessionID]
    if !exists {
        return false
    }
    
    session.Status = status
    return true
}

//возвращает статистику по сессии
func (sm *SessionManager) GetStatistics(sessionID string) (totalRounds, correctAnswers int, percentage float64, ok bool) {
    sm.mu.RLock()
    defer sm.mu.RUnlock()
    
    session, exists := sm.sessions[sessionID]
    if !exists {
        return 0, 0, 0, false
    }
    
    totalRounds = session.TotalRounds
    correctAnswers = session.CorrectAnswers
    if totalRounds > 0 {
        percentage = float64(correctAnswers) / float64(totalRounds) * 100
    }
    return totalRounds, correctAnswers, percentage, true
}