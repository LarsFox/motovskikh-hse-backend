package achievements

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	maxConditionDepth = 8
	maxConditionNodes = 64
)

var ErrInvalidCondition = errors.New("invalid achievement condition")

// Condition is a deliberately small expression tree stored in achievements.condition.
// Exactly one form is allowed per node: all, any, or a field/operator/value predicate.
type Condition struct {
	All      []Condition     `json:"all,omitempty"`
	Any      []Condition     `json:"any,omitempty"`
	Field    string          `json:"field,omitempty"`
	Operator string          `json:"operator,omitempty"`
	Value    json.RawMessage `json:"value,omitempty"`
}

type TestResult struct {
	Event      string
	TestSlug   string
	Percentage float64
	TimeSpent  float64
}

func ParseCondition(raw json.RawMessage) (Condition, error) {
	var condition Condition
	if err := json.Unmarshal(raw, &condition); err != nil {
		return Condition{}, fmt.Errorf("%w: %v", ErrInvalidCondition, err)
	}
	nodes := 0
	if err := validateCondition(condition, 1, &nodes); err != nil {
		return Condition{}, err
	}
	return condition, nil
}

func validateCondition(condition Condition, depth int, nodes *int) error {
	(*nodes)++
	if depth > maxConditionDepth || *nodes > maxConditionNodes {
		return fmt.Errorf("%w: expression is too complex", ErrInvalidCondition)
	}

	groups := 0
	if len(condition.All) > 0 {
		groups++
	}
	if len(condition.Any) > 0 {
		groups++
	}
	predicate := condition.Field != "" || condition.Operator != "" || len(condition.Value) > 0
	if predicate {
		groups++
	}
	if groups != 1 {
		return fmt.Errorf("%w: each node must contain exactly one expression", ErrInvalidCondition)
	}

	for _, child := range append(condition.All, condition.Any...) {
		if err := validateCondition(child, depth+1, nodes); err != nil {
			return err
		}
	}
	if !predicate {
		return nil
	}

	switch condition.Field {
	case "event", "test_slug":
		if condition.Operator != "eq" && condition.Operator != "in" {
			return fmt.Errorf("%w: operator %q is not valid for %s", ErrInvalidCondition, condition.Operator, condition.Field)
		}
	case "percentage", "time_spent":
		if condition.Operator != "eq" && condition.Operator != "gte" && condition.Operator != "lte" {
			return fmt.Errorf("%w: operator %q is not valid for %s", ErrInvalidCondition, condition.Operator, condition.Field)
		}
	default:
		return fmt.Errorf("%w: unsupported field %q", ErrInvalidCondition, condition.Field)
	}
	return nil
}

func Evaluate(condition Condition, result TestResult) (bool, error) {
	if len(condition.All) > 0 {
		for _, child := range condition.All {
			matched, err := Evaluate(child, result)
			if err != nil || !matched {
				return matched, err
			}
		}
		return true, nil
	}
	if len(condition.Any) > 0 {
		for _, child := range condition.Any {
			matched, err := Evaluate(child, result)
			if err != nil {
				return false, err
			}
			if matched {
				return true, nil
			}
		}
		return false, nil
	}

	switch condition.Field {
	case "event":
		return compareString(result.Event, condition.Operator, condition.Value)
	case "test_slug":
		return compareString(result.TestSlug, condition.Operator, condition.Value)
	case "percentage":
		return compareNumber(result.Percentage, condition.Operator, condition.Value)
	case "time_spent":
		return compareNumber(result.TimeSpent, condition.Operator, condition.Value)
	default:
		return false, fmt.Errorf("%w: unsupported field %q", ErrInvalidCondition, condition.Field)
	}
}

func compareString(actual, operator string, raw json.RawMessage) (bool, error) {
	if operator == "eq" {
		var expected string
		if err := json.Unmarshal(raw, &expected); err != nil {
			return false, fmt.Errorf("%w: string value expected", ErrInvalidCondition)
		}
		return actual == expected, nil
	}
	var expected []string
	if err := json.Unmarshal(raw, &expected); err != nil || len(expected) == 0 {
		return false, fmt.Errorf("%w: non-empty string list expected", ErrInvalidCondition)
	}
	for _, value := range expected {
		if actual == value {
			return true, nil
		}
	}
	return false, nil
}

func compareNumber(actual float64, operator string, raw json.RawMessage) (bool, error) {
	var expected float64
	if err := json.Unmarshal(raw, &expected); err != nil {
		return false, fmt.Errorf("%w: number value expected", ErrInvalidCondition)
	}
	switch operator {
	case "eq":
		return actual == expected, nil
	case "gte":
		return actual >= expected, nil
	case "lte":
		return actual <= expected, nil
	default:
		return false, fmt.Errorf("%w: unsupported numeric operator %q", ErrInvalidCondition, operator)
	}
}
