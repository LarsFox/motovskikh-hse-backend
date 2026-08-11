package achievements

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidResult = errors.New("invalid test result")
	ErrUnknownTest   = errors.New("unknown test")
)

type Definition struct {
	ID          int64           `json:"id"`
	Slug        string          `json:"slug"`
	Title       I18n            `json:"title"`
	Description I18n            `json:"description"`
	Condition   json.RawMessage `json:"-"`
	Color       string          `json:"color"`
	SortOrder   int             `json:"sort_order"`
}

type I18n map[string]string

type View struct {
	ID          int64      `json:"id"`
	Slug        string     `json:"slug"`
	Title       I18n       `json:"title"`
	Description I18n       `json:"description"`
	Color       string     `json:"color"`
	Earned      bool       `json:"earned"`
	EarnedAt    *time.Time `json:"earned_at,omitempty"`
}

type Store interface {
	ListActiveDefinitions(ctx context.Context) ([]Definition, error)
	ListUserAwards(ctx context.Context, userID int64) (map[int64]time.Time, error)
	Grant(ctx context.Context, userID, achievementID int64) (bool, time.Time, error)
}

type TestCatalog interface {
	Exists(ctx context.Context, testSlug string) (bool, error)
}

type Service struct {
	store   Store
	catalog TestCatalog
}

func NewService(store Store, catalog TestCatalog) *Service {
	return &Service{store: store, catalog: catalog}
}

func (s *Service) List(ctx context.Context, userID int64, language string) ([]View, error) {
	if userID <= 0 {
		return nil, ErrInvalidResult
	}
	definitions, err := s.store.ListActiveDefinitions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list achievement definitions: %w", err)
	}
	awards, err := s.store.ListUserAwards(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list user achievements: %w", err)
	}
	return makeViews(definitions, awards, language), nil
}

func (s *Service) AwardForTestResult(ctx context.Context, userID int64, result TestResult, language string) ([]View, error) {
	result.Event = "test.completed"
	result.TestSlug = strings.Trim(result.TestSlug, "/ ")
	if userID <= 0 || result.TestSlug == "" || result.Percentage < 0 || result.Percentage > 100 || result.TimeSpent < 0 {
		return nil, ErrInvalidResult
	}
	exists, err := s.catalog.Exists(ctx, result.TestSlug)
	if err != nil {
		return nil, fmt.Errorf("check test catalog: %w", err)
	}
	if !exists {
		return nil, ErrUnknownTest
	}

	definitions, err := s.store.ListActiveDefinitions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list achievement definitions: %w", err)
	}
	newAwards := make(map[int64]time.Time)
	newDefinitions := make([]Definition, 0)
	for _, definition := range definitions {
		condition, err := ParseCondition(definition.Condition)
		if err != nil {
			return nil, fmt.Errorf("achievement %d: %w", definition.ID, err)
		}
		matched, err := Evaluate(condition, result)
		if err != nil {
			return nil, fmt.Errorf("achievement %d: %w", definition.ID, err)
		}
		if !matched {
			continue
		}
		created, awardedAt, err := s.store.Grant(ctx, userID, definition.ID)
		if err != nil {
			return nil, fmt.Errorf("grant achievement %d: %w", definition.ID, err)
		}
		if created {
			newAwards[definition.ID] = awardedAt
			newDefinitions = append(newDefinitions, definition)
		}
	}
	return makeViews(newDefinitions, newAwards, language), nil
}

func makeViews(definitions []Definition, awards map[int64]time.Time, language string) []View {
	views := make([]View, 0, len(definitions))
	for _, definition := range definitions {
		awardedAt, earned := awards[definition.ID]
		view := View{
			ID: definition.ID, Slug: definition.Slug, Color: definition.Color,
			Title: localize(definition.Title, language), Description: localize(definition.Description, language), Earned: earned,
		}
		if earned {
			value := awardedAt.UTC()
			view.EarnedAt = &value
		}
		views = append(views, view)
	}
	return views
}

func localize(value I18n, language string) I18n {
	requestedLanguage := strings.ToLower(language)
	selectedLanguage := "ru"
	if strings.HasPrefix(requestedLanguage, "en") {
		selectedLanguage = "en"
	}
	if translation, ok := value[selectedLanguage]; ok {
		return I18n{selectedLanguage: translation}
	}
	if translation, ok := value["ru"]; ok {
		return I18n{"ru": translation}
	}
	if translation, ok := value["en"]; ok {
		return I18n{"en": translation}
	}
	return I18n{}
}

type StaticCatalog map[string]struct{}

func NewStaticCatalog(slugs []string) StaticCatalog {
	catalog := make(StaticCatalog, len(slugs))
	for _, slug := range slugs {
		catalog[strings.Trim(slug, "/ ")] = struct{}{}
	}
	return catalog
}

func (c StaticCatalog) Exists(_ context.Context, testSlug string) (bool, error) {
	_, ok := c[testSlug]
	return ok, nil
}

var MainPageTests = []string{
	"russia", "russia/cities", "russia/seas", "russia/physical", "russia/rivers",
	"europe", "asia", "africa", "north_america", "south_america", "oceania",
	"world", "world/seas", "globe",
}
