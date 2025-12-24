package entities

// AttemptResult структура - для хранения результатов проверки ответов.
type AttemptResult struct {
    AttemptID   string         `json:"attempt_id"`
    VersionID   string         `json:"version_id"`
    TestID      string         `json:"test_id"`
    UserHash    string         `json:"user_hash"`
    Score       int            `json:"score"`
    MaxScore    int            `json:"max_score"`
    Percentage  float64        `json:"percentage"`
    UserAnswers []*UserAnswer  `json:"user_answers"`
    Details     []QuestionResult `json:"details"`
}

// QuestionResult - детальный результат по вопросу.
type QuestionResult struct {
    QuestionID string `json:"question_id"`
    Text       string `json:"text"`
    UserAnswer string `json:"user_answer"`
    IsCorrect  bool   `json:"is_correct"`
    Score      int    `json:"score"`
    MaxScore   int    `json:"max_score"`
}