package manager

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestValidateAttempt(t *testing.T) {
	m := &Manager{}
	
	tests := []struct {
		name       string
		percentage float64
		timeSpent  int
		wantValid  bool
		wantMsg    string
	}{
		{
			name:       "valid attempt",
			percentage: 70,
			timeSpent:  120,
			wantValid:  true,
			wantMsg:    "",
		},
		{
			name:       "time too short",
			percentage: 70,
			timeSpent:  30,
			wantValid:  false,
			wantMsg:    "Time spent is too short (minimum 60 seconds)",
		},
		{
			name:       "score too low",
			percentage: 3,
			timeSpent:  120,
			wantValid:  false,
			wantMsg:    "Score is too low (minimum 5%)",
		},
		{
			name:       "both invalid",
			percentage: 3,
			timeSpent:  30,
			wantValid:  false,
			wantMsg:    "Time spent is too short (minimum 60 seconds)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, msg := m.ValidateAttempt("test", "user", tt.percentage, tt.timeSpent)
			assert.Equal(t, tt.wantValid, valid)
			assert.Equal(t, tt.wantMsg, msg)
		})
	}
}