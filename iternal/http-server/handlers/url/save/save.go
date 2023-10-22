package save

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"urlShortener/iternal/config"
	res "urlShortener/iternal/pkg/api/response"
	"urlShortener/iternal/pkg/logger/sl"
	"urlShortener/iternal/pkg/random"
	"urlShortener/iternal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	res.Response
	Alias string `json:"alias,omitempty"`
}

type UrlSaver interface {
	SaveURL(UrlToSave string, alias string) error
}

func New(log *slog.Logger, cfg *config.Config, urlSaver UrlSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.http-server.handlers.url.save"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)

		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			render.JSON(w, r, res.Error("empty request"))

			return
		}

		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, res.Error("failed to decode request"))

			return
		}

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, res.ValidationError(validateErr))

			return
		}

		alias := req.Alias

		if alias == "" {
			alias = random.GetRandomStr(cfg.AliasLenght)
		}
		err = urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExist) {
			log.Info("url already exists", slog.String("url", req.URL))

			render.JSON(w, r, res.Error("url already exists"))

			return
		}
		if err != nil {
			log.Error("failed to add url", sl.Err(err))

			render.JSON(w, r, res.Error("failed to add url"))

			return
		}

		render.JSON(w, r, Response{
			Response: res.OK(),
			Alias:    alias,
		})
	}
}
