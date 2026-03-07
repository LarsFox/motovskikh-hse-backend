package entities

type JSON map[string]interface{}

// Test.
type Test struct {
	ID       int    `json:"id" gorm:"column:id;primaryKey;autoIncrement:false"`
	Type     int    `json:"type" gorm:"column:type"`
	Name     string `json:"name" gorm:"column:name;unique"`
	I18n     JSON   `json:"i18n" gorm:"column:i18n;type:json"`
	Settings JSON   `json:"settings" gorm:"column:settings;type:json"`
}
