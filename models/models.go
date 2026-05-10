package models

import "github.com/go-openapi/strfmt"

type Error struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

type GetTestRequest struct {
}


func (r *GetTestRequest) Validate(formats strfmt.Registry) error {
	return nil // Всегда OK для заглушки
}

type GetTestResponse struct {
	Settings *Settings `json:"settings"`
}

type Settings struct {
	CanvasPath string `json:"canvasPath"`
	Hymn       bool   `json:"hymn"`
}

type Task struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Note struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
