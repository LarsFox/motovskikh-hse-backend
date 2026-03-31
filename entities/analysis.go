package entities

// DistributionCategory - категория распределения.
type DistributionCategory struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	MinScore    float64 `json:"min_score"`
	MaxScore    float64 `json:"max_score"`
}


