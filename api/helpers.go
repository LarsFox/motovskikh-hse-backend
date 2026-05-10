package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

const (
	// Бэйс-64 накладывает оверхед примерно треть от файла, 33.33%
	// https://stackoverflow.com/questions/34109053/what-file-size-is-data-if-its-450kb-base64-encoded
	// Максимальный размер устанавливаю 1 МБ * 1.4, размер кодирования бэйс-64 с запасом.
	// В Нжинксе указываю такое же значение.
	maxBodySize = 1468006

	maxNickSize = 42

	playerNameCookie  = "playerReceipe"
	prolongCookieTime = time.Hour * 24 * 365 // 1 год.
)

// Внутренние ошибки.
var (
	errUnknownError = errors.New("unknown error")
)

func cutNick(nick string) string {
	var i int
	for j := range nick {
		if i == maxNickSize {
			return nick[:j]
		}
		i++
	}

	return nick
}

func getPlayerHeader(r *http.Request) (http.Header, string, error) {
	cookie, err := r.Cookie(playerNameCookie)
	switch {
	case errors.Is(err, http.ErrNoCookie):
		header := http.Header{}
		player := entities.NewPlayerName()

		cookie := &http.Cookie{
			Name:     playerNameCookie,
			Value:    player,
			Expires:  time.Now().Add(prolongCookieTime),
			HttpOnly: true,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
			Secure:   true,
		}

		header.Add("Set-Cookie", cookie.String())
		return header, player, nil

	case errors.Is(err, nil):
		return http.Header{}, cookie.Value, nil

	default:
		return nil, "", err
	}
}

func notify(e error, meta ...map[string]any) {
	if strings.Contains(e.Error(), "write: broken pipe") {
		return
	}

	if errors.Is(e, context.Canceled) {
		return
	}

	entities.Notify(e, meta...)
}

func notifyRecover(meta ...map[string]any) {
	rec := recover()
	if rec == nil {
		return
	}

	var err error
	switch t := rec.(type) {
	case error:
		err = t
	default:
		err = errUnknownError
	}

	meta = append(meta, map[string]any{
		"stack": string(debug.Stack()),
	})

	notify(err, meta...)
}

// unmarshalParams парсит входящие в методы API параметры и валидирует их.
func unmarshalParams(r *http.Request, prms runtime.Validatable) error {
	if err := json.NewDecoder(r.Body).Decode(prms); err != nil {
		return err
	}

	if err := r.Body.Close(); err != nil {
		return err
	}

	if err := prms.Validate(strfmt.Default); err != nil {
		return err
	}

	return nil
}
