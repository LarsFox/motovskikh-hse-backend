package mysql

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm/clause"

	"github.com/LarsFox/motovskikh-hse-backend/achievements"
)

type achievement struct {
	ID            int64           `gorm:"primaryKey;autoIncrement"`
	Slug          string          `gorm:"size:96;uniqueIndex;not null"`
	TitleRU       string          `gorm:"size:160;not null"`
	TitleEN       string          `gorm:"size:160;not null"`
	DescriptionRU string          `gorm:"size:500;not null"`
	DescriptionEN string          `gorm:"size:500;not null"`
	Condition     json.RawMessage `gorm:"type:json;not null"`
	Color         string          `gorm:"size:7;not null"`
	IsActive      bool            `gorm:"not null;default:true;index"`
	SortOrder     int             `gorm:"not null;default:0"`
}

type userAchievement struct {
	UserID        int64     `gorm:"primaryKey;autoIncrement:false"`
	AchievementID int64     `gorm:"primaryKey;autoIncrement:false;index"`
	AwardedAt     time.Time `gorm:"not null"`
}

func (c *Client) ListActiveDefinitions(ctx context.Context) ([]achievements.Definition, error) {
	rows := make([]achievement, 0)
	if err := c.db.WithContext(ctx).Where("is_active = ?", true).Order("sort_order, id").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list achievements: %w", err)
	}
	result := make([]achievements.Definition, 0, len(rows))
	for _, row := range rows {
		result = append(result, achievements.Definition{
			ID: row.ID, Slug: row.Slug, TitleRU: row.TitleRU, TitleEN: row.TitleEN,
			DescriptionRU: row.DescriptionRU, DescriptionEN: row.DescriptionEN,
			Condition: row.Condition, Color: row.Color, SortOrder: row.SortOrder,
		})
	}
	return result, nil
}

func (c *Client) ListUserAwards(ctx context.Context, userID int64) (map[int64]time.Time, error) {
	rows := make([]userAchievement, 0)
	if err := c.db.WithContext(ctx).Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list user achievements: %w", err)
	}
	result := make(map[int64]time.Time, len(rows))
	for _, row := range rows {
		result[row.AchievementID] = row.AwardedAt
	}
	return result, nil
}

func (c *Client) Grant(ctx context.Context, userID, achievementID int64) (bool, time.Time, error) {
	awardedAt := time.Now().UTC()
	row := &userAchievement{UserID: userID, AchievementID: achievementID, AwardedAt: awardedAt}
	result := c.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(row)
	if result.Error != nil {
		return false, time.Time{}, fmt.Errorf("grant achievement: %w", result.Error)
	}
	return result.RowsAffected == 1, awardedAt, nil
}
