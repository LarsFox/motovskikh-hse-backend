package entities

import "time"

// DistributionCategory - категория распределения.
type DistributionCategory struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	MinScore    float64 `json:"min_score"`
	MaxScore    float64 `json:"max_score"`
}

// PerformanceQuadrant - квадрант производительности.
type PerformanceQuadrant struct {
	Quadrant    string  `json:"quadrant"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	XPosition   float64 `json:"x_position"`
	YPosition   float64 `json:"y_position"`
}

// DetailedAnalysis - детальный анализ попытки.
type DetailedAnalysis struct {
	AttemptID     string                 `json:"attempt_id"`
	TestID        string                 `json:"test_id"`
	UserHash      string                 `json:"user_hash,omitempty"`
	
	Score         int                    `json:"score"`
	MaxScore      int                    `json:"max_score"`
	Percentage    float64                `json:"percentage"`
	TimeSpent     int                    `json:"time_spent"`
	CreatedAt     time.Time              `json:"created_at"`
	
	PercentileRank   float64                `json:"percentile_rank"`
	TimePercentile   float64                `json:"time_percentile"`
	DistributionCategory *DistributionCategory  `json:"distribution_category"`
	SkillLevel          string                 `json:"skill_level"`
	
	TestStats          *TestStats              `json:"test_stats,omitempty"`
	ByQuestionType    map[string]QuestionTypeStats `json:"by_question_type"`
	Recommendations   []Recommendation           `json:"recommendations,omitempty"`
}

// QuestionTypeStats - статистика по типу вопросов.
type QuestionTypeStats struct {
	TotalQuestions  int     `json:"total_questions"`
	CorrectAnswers  int     `json:"correct_answers"`
	Percentage      float64 `json:"percentage"`
	PointsEarned    int     `json:"points_earned"`
	MaxPoints       int     `json:"max_points"`
}

// Recommendation - рекомендация для улучшения.
type Recommendation struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"`
	QuestionTypes []string `json:"question_types,omitempty"`
}
