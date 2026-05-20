package entities

import (
	"encoding/json"
	"time"
)

type WorkshopTag struct {
	ID        int64
	Name      string
	I18n      map[string]string
	Total     int64
	ParentIDs string
	Joined    json.RawMessage // Чтобы парсить единожды на выдаче.
}

type WorkshopTest struct {
	ID        int64
	Name      string
	Secret    string
	Mark      string
	Official  string
	Likes     int64
	Themes    []string
	Ban       string
	Meta      *WorkshopTestMeta
	Src       *WorkshopTestSrc
	CreatedAt time.Time
	Tag       *WorkshopTag // для сджоенного запроса.
}

type WorkshopTestMeta struct {
	Author    string `json:"author"`
	Image     string `json:"image,omitempty"`
	Questions int64  `json:"questions"`
	Title     string `json:"title"`
}

type WorkshopTestSrc struct {
	Objects  []*MapObject   `json:"objects"`
	Settings *TestSettings  `json:"settings"`
	Texts    map[string]any `json:"texts"`
	TextsEn  map[string]any `json:"texts_en,omitempty"`
}
