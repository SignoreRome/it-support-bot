package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"net/http"
	"support_bot/env"
	"sync"
)

type RestartRequest struct {
	ChatId int64  `json:"chatId"`
	Stand  string `json:"stand"`
}

func (u *RestartRequest) Bind(r *http.Request) error {
	return nil
}

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

func main() {

	wg := sync.WaitGroup{}
	wg.Add(1)

	apiToken := env.GetString("TELEGRAM_API_TOKEN", "")
	if apiToken == "" {
		panic("TELEGRAM_API_TOKEN must be passed")
	}
	bot, err := tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		panic(err)
	}
	bot.Debug = true

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/restart", func(w http.ResponseWriter, r *http.Request) {
		data := &RestartRequest{}
		if err := render.Bind(r, data); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}
		chatId := data.ChatId
		stand := data.Stand
		msg := tgbotapi.NewMessage(chatId, "Запущено обновление стенда "+stand)
		if _, err := bot.Send(msg); err != nil {
			return
		}
	})

	err = http.ListenAndServe(":3333", r)
	if err != nil {
		panic(err)
	}

}
