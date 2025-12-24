package mysql

import (
	"math/rand"
	"time"
	"log"
	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/google/uuid"
)

// Создание тестовых данных. (Generated)
func (c *Client) CreateTestData() error {
	// Очищаем все данные полностью
	c.db.Exec("DELETE FROM user_answers")
	c.db.Exec("DELETE FROM attempts")
	c.db.Exec("DELETE FROM question_stats")
	c.db.Exec("DELETE FROM test_stats")
	c.db.Exec("DELETE FROM questions")
	c.db.Exec("DELETE FROM tests")
	c.db.Exec("DELETE FROM test_configs")
	c.db.Exec("DELETE FROM test_versions")
	
	tests := []*entities.Test{
		{
			TestID:      "geography_basic",
			Title:       "География: основы",
			Description: "Простые вопросы по географии",
			Category:    "География",
			Difficulty:  2,
			CreatedBy:   "system",
			IsPublic:    true,
			CreatedAt:   time.Now(),
		},
		{
			TestID:      "math_basic",
			Title:       "Математика: основы",
			Description: "Простые математические задачи",
			Category:    "Математика",
			Difficulty:  2,
			CreatedBy:   "system",
			IsPublic:    true,
			CreatedAt:   time.Now(),
		},
	}
	
	// Создаем вопросы для тестов
	questions := []*entities.Question{
		// География
		{
			QuestionID:   "geo_q1",
			TestID:       "geography_basic",
			Text:         "Столица России?",
			QuestionType: "single",
			Options:      `["Москва", "Санкт-Петербург", "Казань", "Новосибирск"]`,
			CorrectAnswer: `"Москва"`,
			Points:       10,
			OrderIndex:   1,
			CreatedAt:    time.Now(),
		},
		{
			QuestionID:   "geo_q2",
			TestID:       "geography_basic",
			Text:         "Какие страны находятся в Европе?",
			QuestionType: "multiple",
			Options:      `["Франция", "Германия", "Бразилия", "Италия", "Япония"]`,
			CorrectAnswer: `["Франция", "Германия", "Италия"]`,
			Points:       15,
			OrderIndex:   2,
			CreatedAt:    time.Now(),
		},
		// Математика
		{
			QuestionID:   "math_q1",
			TestID:       "math_basic",
			Text:         "3 × 4 = ?",
			QuestionType: "single",
			Options:      `["7", "12", "15", "9"]`,
			CorrectAnswer: `"12"`,
			Points:       10,
			OrderIndex:   1,
			CreatedAt:    time.Now(),
		},
		{
			QuestionID:   "math_q2",
			TestID:       "math_basic",
			Text:         "Простое число?",
			QuestionType: "multiple",
			Options:      `["2", "4", "7", "9", "11"]`,
			CorrectAnswer: `["2", "7", "11"]`,
			Points:       15,
			OrderIndex:   2,
			CreatedAt:    time.Now(),
		},
	}
	
	// Создаем версии тестов
	versions := []*entities.TestVersion{
		{
			VersionID:     "geo_v1",
			TestID:        "geography_basic",
			VersionNumber: 1,
			Description:   "Первая версия",
			CreatedAt:     time.Now(),
			IsActive:      true,
		},
		{
			VersionID:     "math_v1",
			TestID:        "math_basic",
			VersionNumber: 1,
			Description:   "Первая версия",
			CreatedAt:     time.Now(),
			IsActive:      true,
		},
	}
	
	// Создаем конфигурации тестов
	configs := []*entities.TestConfig{
		{
			ConfigID:           uuid.New().String(),
			TestID:             "geography_basic",
			MinTimeSpent:       30,
			MaxTimeSpent:       600,
			MinPercentage:      10,
			MaxAttemptsPerUser: 5,
			UpdatedAt:          time.Now(),
		},
		{
			ConfigID:           uuid.New().String(),
			TestID:             "math_basic",
			MinTimeSpent:       20,
			MaxTimeSpent:       300,
			MinPercentage:      20,
			MaxAttemptsPerUser: 5,
			UpdatedAt:          time.Now(),
		},
	}
	
	// Сохраняем все в БД
	for _, test := range tests {
		if err := c.db.Create(test).Error; err != nil {
			// Игнорируем дублирование
			continue
		}
	}
	
	for _, question := range questions {
		if err := c.db.Create(question).Error; err != nil {
			continue
		}
	}
	
	for _, version := range versions {
		if err := c.db.Create(version).Error; err != nil {
			continue
		}
	}
	
	for _, config := range configs {
		if err := c.db.Create(config).Error; err != nil {
			continue
		}
	}
	
	// Создаем реалистичные попытки.
	rand.Seed(time.Now().UnixNano())
	
	// Попытки для географии.
	for i := 0; i < 15; i++ {
		// Реалистичное распределение: большинство в районе 60-80%
		baseScore := 60.0 + rand.Float64()*20.0
		// Добавляем случайность
		percentage := baseScore + (rand.Float64()*20 - 10)
		if percentage < 0 {
			percentage = 0
		}
		if percentage > 100 {
			percentage = 100
		}
		
		score := int(percentage * 25 / 100) // Максимум 25 баллов за тест
		maxScore := 25
		
		attempt := &entities.Attempt{
			ID:         uuid.New().String(),
			TestID:     "geography_basic",
			VersionID:  "geo_v1",
			UserHash:   generateUserHash(i),
			Score:      score,
			MaxScore:   maxScore,
			Percentage: float64(score) / float64(maxScore) * 100,
			TimeSpent:  60 + rand.Intn(240), // 1-5 минут
			Answers:    generateGeographyAnswers(),
			IsValid:    true,
			CreatedAt:  time.Now().Add(-time.Duration(rand.Intn(168)) * time.Hour), // до 7 дней назад
		}
		
		if err := c.db.Create(attempt).Error; err != nil {
			return err
		}
	}
	
	// Попытки для математики (15 штук)
	for i := 0; i < 15; i++ {
		// Для математики результаты обычно выше
		baseScore := 70.0 + rand.Float64()*25.0
		percentage := baseScore + (rand.Float64()*15 - 7.5)
		if percentage < 0 {
			percentage = 0
		}
		if percentage > 100 {
			percentage = 100
		}
		
		score := int(percentage * 25 / 100)
		maxScore := 25
		
		attempt := &entities.Attempt{
			ID:         uuid.New().String(),
			TestID:     "math_basic",
			VersionID:  "math_v1",
			UserHash:   generateUserHash(i + 15),
			Score:      score,
			MaxScore:   maxScore,
			Percentage: float64(score) / float64(maxScore) * 100,
			TimeSpent:  30 + rand.Intn(150), // 0.5-3 минуты
			Answers:    generateMathAnswers(),
			IsValid:    true,
			CreatedAt:  time.Now().Add(-time.Duration(rand.Intn(120)) * time.Hour), // до 5 дней назад
		}
		
		if err := c.db.Create(attempt).Error; err != nil {
			return err
		}
	}
	
	// Теперь вызываем автоматический расчет статистики
	for _, test := range tests {
		// Расчет общей статистики теста
		if err := c.CalculateTestStats(test.TestID); err != nil {
			// Логируем ошибку, но не прерываем
			log.Printf("Ошибка расчета статистики для теста %s: %v", test.TestID, err)
		}
		
		// Расчет статистики по вопросам
		versionID := getVersionIDForTest(test.TestID, versions)
		if err := c.CalculateQuestionStats(test.TestID, versionID); err != nil {
			log.Printf("Ошибка расчета статистики вопросов для теста %s: %v", test.TestID, err)
		}
	}
	
	log.Println("Тестовые данные созданы, статистика рассчитана автоматически")
	return nil
}

func generateUserHash(index int) string {
	users := []string{
		"user_alex", "user_maria", "user_dmitry", "user_olga", "user_sergey",
		"user_anna", "user_mikhail", "user_ekaterina", "user_andrey", "user_natalia",
		"user_vladimir", "user_irina", "user_pavel", "user_tatyana", "user_ivan",
		"user_elena", "user_nikolay", "user_svetlana", "user_yuri", "user_larisa",
	}
	
	if index < len(users) {
		return users[index]
	}
	return "user_" + uuid.New().String()[:6]
}

func getVersionIDForTest(testID string, versions []*entities.TestVersion) string {
	for _, v := range versions {
		if v.TestID == testID {
			return v.VersionID
		}
	}
	return "v1"
}

func generateGeographyAnswers() string {
	answers := []string{
		`{"geo_q1": "Москва", "geo_q2": "[\"Франция\", \"Германия\", \"Италия\"]"}`,
		`{"geo_q1": "Москва", "geo_q2": "[\"Франция\", \"Германия\"]"}`,
		`{"geo_q1": "Санкт-Петербург", "geo_q2": "[\"Франция\", \"Германия\", \"Италия\"]"}`,
		`{"geo_q1": "Москва", "geo_q2": "[\"Франция\", \"Бразилия\", \"Италия\"]"}`,
		`{"geo_q1": "Москва", "geo_q2": "[\"Германия\", \"Италия\", \"Япония\"]"}`,
	}
	return answers[rand.Intn(len(answers))]
}

func generateMathAnswers() string {
	answers := []string{
		`{"math_q1": "12", "math_q2": "[\"2\", \"7\", \"11\"]"}`,
		`{"math_q1": "12", "math_q2": "[\"2\", \"7\"]"}`,
		`{"math_q1": "15", "math_q2": "[\"2\", \"7\", \"11\"]"}`,
		`{"math_q1": "12", "math_q2": "[\"4\", \"7\", \"11\"]"}`,
		`{"math_q1": "12", "math_q2": "[\"2\", \"9\", \"11\"]"}`,
	}
	return answers[rand.Intn(len(answers))]
}