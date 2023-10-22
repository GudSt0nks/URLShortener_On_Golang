package get

import (
	"errors"
	"log/slog"
	"net/http"
	"urlShortener/iternal/config"
	res "urlShortener/iternal/pkg/api/response"
	"urlShortener/iternal/pkg/logger/sl"
	"urlShortener/iternal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	res.Response
	Url string `json:"url,omitempty"`
}

type UrlContainer interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, cfg *config.Config, urlContainer UrlContainer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.http-server.handlers.url.get"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")

		if alias == "" {
			log.Info("url is empty")

			render.JSON(w, r, res.Error("Alias is required"))
		}

		resURL, err := urlContainer.GetURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)

			render.JSON(w, r, res.Error("not found"))

			return
		}
		if err != nil {
			log.Error("failed to get url", sl.Err(err))

			render.JSON(w, r, res.Error("internal error"))

			return
		}

		log.Info("got url", slog.String("url", resURL))

		http.Redirect(w, r, resURL, http.StatusFound)
	}
}
