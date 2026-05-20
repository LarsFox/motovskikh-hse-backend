package entities

// MapObject — объединенный с областью объект.
type MapObject struct {
	Capital  *Object `json:"Capital,omitempty"`
	Name     string
	Object   *Object
	Params   map[string]any `json:"Params,omitempty"`
	Src      string
	Subclass string `json:"Subclass,omitempty"`
}

type Object struct {
	I18n   map[string]string
	Params map[string]any `json:"Params,omitempty"`
}
