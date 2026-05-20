package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

const workshopListUserTestsAlloc = 100

const queryListUserTests = `
	SELECT
		tests.id,
		tests.name,
		tests.mark,
		tests.official,
		tests.likes,
		tests.created_at,
		tests.meta
	FROM workshop_tests tests
	WHERE tests.user_id = ?
		AND tests.ban IS NULL`

type workshopTest struct {
	ID        int64
	Name      string
	Secret    string
	Src       json.RawMessage
	Meta      json.RawMessage
	Mark      sql.NullString
	Official  sql.NullString
	Likes     int64
	Ban       sql.NullString
	CreatedAt time.Time
}

func (c *Client) WorkshopListUserTests(ctx context.Context, userID int64) ([]*entities.WorkshopTest, error) {
	defer c.track("WorkshopListUserTests", "", time.Now())

	wts := make([]*workshopTest, 0, workshopListUserTestsAlloc)
	if err := c.db.WithContext(ctx).
		Raw(queryListUserTests, userID).
		Find(&wts).Error; err != nil {
		return nil, fmt.Errorf("workshop list user tests err: %w", err)
	}

	result := make([]*entities.WorkshopTest, 0, len(wts))
	for _, wt := range wts {
		test := &entities.WorkshopTest{
			ID:        wt.ID,
			Name:      wt.Name,
			Mark:      wt.Mark.String,
			Official:  wt.Official.String,
			Likes:     wt.Likes,
			CreatedAt: wt.CreatedAt,
		}

		if err := json.Unmarshal(wt.Meta, &test.Meta); err != nil {
			return nil, fmt.Errorf("workshop list user tests unmarshalling meta err: %w", err)
		}

		result = append(result, test)
	}

	return result, nil
}
