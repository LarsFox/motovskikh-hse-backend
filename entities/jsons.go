package entities

// TestSettings — все возможные настройки теста.
//
// По-хорошему должны подаваться без Джейсон-тегов и маршаллиться на уровне хранилища без них.
// По чистой архитектуре ведь никому нет дела, как эти настройки сохраняются.
// Но неохота обновлять всю базу ради чистоты, поэтому из-за легаси оставляю с ними.
type TestSettings struct {
	Boxies        int64    `json:"boxies,omitempty"`        // ширина коробочки для старых тестов.
	CanvasPath    string   `json:"canvas_path,omitempty"`   // путь до холста.
	CSSLanguages  []string `json:"css_languages,omitempty"` // дополнительные языки в тесте как CSS файлы: fa требует ar, de не требует ничего.
	CustomCSS     []string `json:"custom_css,omitempty"`    // дополнительные стили.
	Dev           bool     `json:"dev,omitempty"`           // разрабатывается ли тест, false — тест доступен всем.
	Height        int64    `json:"height"`                  // высота изображения.
	Hard          bool     `json:"hard,omitempty"`          // усложненный ли тест.
	Hymn          bool     `json:"hymn,omitempty"`          // доступен ли гимн.
	International bool     `json:"international,omitempty"` // доступна ли версия на английском.
	Languages     []string `json:"languages,omitempty"`     // дополнительные языки в тесте.
	Modes         []string `json:"modes,omitempty"`         // доступные режимы игры.
	Subjects      string   `json:"subjects,omitempty"`      // тип субъктов, влияет на текст в настройках.
	Width         int64    `json:"width"`                   // ширина изображения.
	Zoom          int64    `json:"zoom,omitempty"`          // допустимый максимум зума.
}

// TestI18n — именование теста.
type TestI18n map[string]*TestI18nLoc

// TestI18nLoc — словарь строк для локализации теста.
type TestI18nLoc struct {
	Acc  string `json:"acc"`  // название субъекта в винительном падеже.
	Desc string `json:"desc"` // что ищем: субъектов Германии, вилайетов Алжира, мухафаз Египта.
}
