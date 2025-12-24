package mysql

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/google/uuid"
)

// Config — конфигурация клиента.
type Config struct {
	Host    string `envconfig:"optional"`
	Pass    string
	MaxConn int `envconfig:"default=0"`
	Name    string
	User    string
}

type Client struct {
	db *gorm.DB // GORM ORM для работы с БД.
}

// Подключение.
func (c *Config) connection() string {
	sqlHost := c.Host
	if !strings.Contains(sqlHost, "tcp") {
		sqlHost = fmt.Sprintf("tcp(%s)", sqlHost)
	}
	return fmt.Sprintf("%s:%s@%s/%s?charset=utf8mb4&parseTime=True&loc=Local", c.User, c.Pass, sqlHost, c.Name)
}

func NewClient(cfg *Config) (*Client, error) {
	db, err := gorm.Open(mysql.Open(cfg.connection()))
	if err != nil {
		return nil, fmt.Errorf("dbs new client err: %w", err)
	}

	// Автоматическое создание всех таблиц.
	err = db.AutoMigrate(
		&entities.Test{},
		&entities.Question{},
		&entities.TestConfig{},
		&entities.Attempt{},
		&entities.UserAnswer{},
		&entities.TestStats{},
		&entities.QuestionStats{},
		&entities.TestVersion{},
	)
	if err != nil {
		return nil, fmt.Errorf("dbs migrate err: %w", err)
	}

	d, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("dbs new client err: %w", err)
	}

	d.SetMaxOpenConns(cfg.MaxConn)
	return &Client{db: db}, nil
}

// CreateTest создает новый тест.
func (c *Client) CreateTest(test *entities.Test, questions []*entities.Question) error {
    return c.db.Transaction(func(tx *gorm.DB) error {
        // Создаем тест.
        if err := tx.Create(test).Error; err != nil {
            return err
        }
        // Создаем вопросы.
        for _, q := range questions {
            q.TestID = test.TestID
            if err := tx.Create(q).Error; err != nil {
                return err
            }
        }
        // Создаем первую версию.
        version := &entities.TestVersion{
            VersionID:     uuid.New().String(),
            TestID:        test.TestID,
            VersionNumber: 1,
            Description:   "Первая версия теста",
            CreatedAt:     time.Now(),
            IsActive:      true,
        }
        if err := tx.Create(version).Error; err != nil {
            return err
        }
        // Создаем конфигурацию по умолчанию.
        config := &entities.TestConfig{
            ConfigID:           uuid.New().String(),
            TestID:             test.TestID,
            MinTimeSpent:       30,
            MaxTimeSpent:       3600,
            MinPercentage:      0,
            MaxAttemptsPerUser: 3,
            UpdatedAt:          time.Now(),
        }
        return tx.Create(config).Error
    })
}

// GetTest возвращает тест с вопросами.
func (c *Client) GetTest(testID string) (*entities.Test, []*entities.Question, error) {
	var test entities.Test
	if err := c.db.First(&test, "test_id = ?", testID).Error; err != nil {
		return nil, nil, err
	}
	
	var questions []*entities.Question
	if err := c.db.Where("test_id = ?", testID).
		Order("order_index ASC").
		Find(&questions).Error; err != nil {
		return &test, nil, err
	}
	
	// Скрываем правильные ответы.
	for _, q := range questions {
		q.CorrectAnswer = ""
	}
	
	return &test, questions, nil
}

// CheckAnswers проверяет ответы и возвращает результат.
func (c *Client) CheckAnswers(testID string, userHash string, answers map[string]string) (*entities.AttemptResult, error) {
    // Получаем активную версию теста.
    versionID, err := c.GetActiveVersionID(testID)
    if err != nil {
        versionID = ""
    }
    // Получаем вопросы с правильными ответами.
    var questions []*entities.Question
    if err := c.db.Where("test_id = ?", testID).
        Order("order_index ASC").
        Find(&questions).Error; err != nil {
        return nil, err
    }
    // Проверяем каждый ответ.
    totalScore := 0
    maxScore := 0
    userAnswers := make([]*entities.UserAnswer, 0, len(questions))
    detailedResults := make([]entities.QuestionResult, 0, len(questions))
    
    for _, q := range questions {
        maxScore += q.Points
        
        userAnswer, ok := answers[q.QuestionID]
        if !ok {
            // Пользователь не ответил на вопрос.
            userAnswers = append(userAnswers, &entities.UserAnswer{
                AnswerID:   uuid.New().String(),
                QuestionID: q.QuestionID,
                UserHash:   userHash,
                Answer:     "",
                IsCorrect:  false,
                Score:      0,
                CreatedAt:  time.Now(),
            })
            continue
        }
        
        // Проверяем правильность ответа.
        isCorrect := false
        var correctAnswers []string
        
        if err := json.Unmarshal([]byte(q.CorrectAnswer), &correctAnswers); err != nil {
            // Если не JSON, значит текстовый ответ.
            isCorrect = strings.EqualFold(strings.TrimSpace(userAnswer), 
                strings.TrimSpace(q.CorrectAnswer))
        } else {
            // Для multiple choice сравниваем JSON.
            var userChoices []string
            json.Unmarshal([]byte(userAnswer), &userChoices)

            isCorrect = compareStringSlices(userChoices, correctAnswers)
        }
        score := 0
        if isCorrect {
            score = q.Points
            totalScore += score
        }
        // Сохраняем детальный ответ.
        userAnswers = append(userAnswers, &entities.UserAnswer{
            AnswerID:   uuid.New().String(),
            QuestionID: q.QuestionID,
            UserHash:   userHash,
            Answer:     userAnswer,
            IsCorrect:  isCorrect,
            Score:      score,
            CreatedAt:  time.Now(),
        })
        
        // Подробные результаты.
        detailedResults = append(detailedResults, entities.QuestionResult{
            QuestionID: q.QuestionID,
            Text:       q.Text,
            UserAnswer: userAnswer,
            IsCorrect:  isCorrect,
            Score:      score,
            MaxScore:   q.Points,
        })
    }
    
    percentage := float64(totalScore) / float64(maxScore) * 100
    
    return &entities.AttemptResult{
        TestID:       testID,
        VersionID:    versionID,
        UserHash:     userHash,
        Score:        totalScore,
        MaxScore:     maxScore,
        Percentage:   percentage,
        UserAnswers:  userAnswers,
        Details:      detailedResults,
    }, nil
}

// Вспомогательная функция для сравнения строк.
func compareStringSlices(a, b []string) bool {
    if len(a) != len(b) {
        return false
    }
    // Создаем map для сравнения без учета порядка.
    mapA := make(map[string]int)
    mapB := make(map[string]int)
    for _, val := range a {
        mapA[val]++
    }
    for _, val := range b {
        mapB[val]++
    }
    // Сравниваем.
    if len(mapA) != len(mapB) {
        return false
    }
    for key, countA := range mapA {
        if countB, ok := mapB[key]; !ok || countB != countA {
            return false
        }
    }
    return true
}
// SaveFullAttempt сохраняет полную попытку с ответами.
func (c *Client) SaveFullAttempt(attempt *entities.Attempt, userAnswers []*entities.UserAnswer) error {
	return c.db.Transaction(func(tx *gorm.DB) error {
		// Сохраняем основную попытку.
		if err := tx.Create(attempt).Error; err != nil {
			return err
		}
		// Сохраняем ответы на вопросы.
		for _, answer := range userAnswers {
			answer.AttemptID = attempt.ID
			if err := tx.Create(answer).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// CreateNewVersion создает новую версию теста.
func (c *Client) CreateNewVersion(testID, description string) (string, error) {
    // Находим текущую максимальную версию.
    var maxVersion int
    err := c.db.Model(&entities.TestVersion{}).
        Where("test_id = ?", testID).
        Select("COALESCE(MAX(version_number), 0)").
        Scan(&maxVersion).Error
    if err != nil {
        return "", err
    }
    
    // Деактивируем старые версии.
    err = c.db.Model(&entities.TestVersion{}).
        Where("test_id = ?", testID).
        Update("is_active", false).Error
    if err != nil {
        return "", err
    }
    
    // Создаем новую версию.
    version := &entities.TestVersion{
        VersionID:     uuid.New().String(),
        TestID:        testID,
        VersionNumber: maxVersion + 1,
        Description:   description,
        CreatedAt:     time.Now(),
        IsActive:      true,
    }
    return version.VersionID, c.db.Create(version).Error
}

// GetActiveVersionID возвращает ID активной версии теста.
func (c *Client) GetActiveVersionID(testID string) (string, error) {
    var version entities.TestVersion
    err := c.db.Where("test_id = ? AND is_active = ?", testID, true).
        First(&version).Error
    if err != nil {
        return "", err
    }
    return version.VersionID, nil
}
