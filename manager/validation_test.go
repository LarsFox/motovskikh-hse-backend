package manager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateAttempt_WithTest(t *testing.T) {
	m := &Manager{}

	tests := []struct {
		name          string
		testName      string
		percentage    float64
		timeSpent     float64
		questionCount int64
		expectValid   bool
	}{
		{
			name:          "valid attempt with 30 regions",
			testName:      "europe",
			percentage:    70,
			timeSpent:     180,
			questionCount: 30,
			expectValid:   true,
		},
		{
			name:          "too fast for 30 regions",
			testName:      "europe",
			percentage:    70,
			timeSpent:     30,
			questionCount: 30,
			expectValid:   false,
		},
		{
			name:          "too low percentage",
			testName:      "europe",
			percentage:    3,
			timeSpent:     180,
			questionCount: 30,
			expectValid:   false,
		},
		{
			name:          "small test lower boundary",
			testName:      "small",
			percentage:    8,
			timeSpent:     60,
			questionCount: 8,
			expectValid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := m.validateAttempt(
				tt.testName,
				tt.percentage,
				tt.timeSpent,
				tt.questionCount,
			)

			assert.Equal(t, tt.expectValid, valid)
		})
	}
}