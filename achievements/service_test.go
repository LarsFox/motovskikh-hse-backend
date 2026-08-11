package achievements

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type memoryStore struct {
	definitions []Definition
	awards      map[int64]map[int64]time.Time
}

func (s *memoryStore) ListActiveDefinitions(context.Context) ([]Definition, error) {
	return s.definitions, nil
}

func (s *memoryStore) ListUserAwards(_ context.Context, userID int64) (map[int64]time.Time, error) {
	result := map[int64]time.Time{}
	for id, awardedAt := range s.awards[userID] {
		result[id] = awardedAt
	}
	return result, nil
}

func (s *memoryStore) Grant(_ context.Context, userID, achievementID int64) (bool, time.Time, error) {
	if s.awards[userID] == nil {
		s.awards[userID] = map[int64]time.Time{}
	}
	if awardedAt, ok := s.awards[userID][achievementID]; ok {
		return false, awardedAt, nil
	}
	awardedAt := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	s.awards[userID][achievementID] = awardedAt
	return true, awardedAt, nil
}

func definition(id int64, slug, condition string) Definition {
	return Definition{
		ID: id, Slug: slug,
		Title: I18n{"ru": "Рус", "en": "English"}, Description: I18n{"ru": "Описание", "en": "Description"},
		Condition: json.RawMessage(condition), Color: "#ffbf69", SortOrder: int(id),
	}
}

func TestAwardForTestResultIsValidatedAndIdempotent(t *testing.T) {
	store := &memoryStore{
		definitions: []Definition{
			definition(1, "first-mark", `{"all":[{"field":"event","operator":"eq","value":"test.completed"},{"field":"test_slug","operator":"in","value":["russia","europe"]}]}`),
			definition(2, "perfect-map", `{"all":[{"field":"test_slug","operator":"eq","value":"russia"},{"field":"percentage","operator":"gte","value":100}]}`),
		},
		awards: map[int64]map[int64]time.Time{},
	}
	service := NewService(store, NewStaticCatalog([]string{"russia", "europe"}))

	awarded, err := service.AwardForTestResult(context.Background(), 42, TestResult{TestSlug: "/russia/", Percentage: 100, TimeSpent: 12}, "en-US")
	if err != nil || len(awarded) != 2 || len(awarded[0].Title) != 1 || awarded[0].Title["en"] != "English" {
		t.Fatalf("unexpected first award: %#v, %v", awarded, err)
	}
	repeated, err := service.AwardForTestResult(context.Background(), 42, TestResult{TestSlug: "russia", Percentage: 100}, "en")
	if err != nil || len(repeated) != 0 {
		t.Fatalf("award must be idempotent: %#v, %v", repeated, err)
	}
	all, err := service.List(context.Background(), 42, "ru")
	if err != nil || len(all) != 2 || !all[0].Earned || len(all[0].Title) != 1 || all[0].Title["ru"] != "Рус" {
		t.Fatalf("unexpected catalog: %#v, %v", all, err)
	}
}

func TestAwardRejectsUnknownTest(t *testing.T) {
	service := NewService(&memoryStore{awards: map[int64]map[int64]time.Time{}}, NewStaticCatalog([]string{"russia"}))
	_, err := service.AwardForTestResult(context.Background(), 42, TestResult{TestSlug: "secret", Percentage: 50}, "ru")
	if !errors.Is(err, ErrUnknownTest) {
		t.Fatalf("expected ErrUnknownTest, got %v", err)
	}
}

func TestConditionRejectsUnsupportedFields(t *testing.T) {
	_, err := ParseCondition(json.RawMessage(`{"field":"sql","operator":"eq","value":"DROP"}`))
	if !errors.Is(err, ErrInvalidCondition) {
		t.Fatalf("expected ErrInvalidCondition, got %v", err)
	}
}
