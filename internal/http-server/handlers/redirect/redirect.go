package redirect

import (
	// project
	"go-url-shortener/internal/lib/api/response"
	"go-url-shortener/internal/lib/logger/sl"
	"go-url-shortener/internal/storage"

	// embedded
	"errors"
	"log/slog"
	"net/http"

	// external
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)


//go:generate go run github.com/vektra/mockery/v2@latest --name=URLGetter --output=mocks --outpkg=mocks --with-expecter
type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")
			render.JSON(w, r, response.Error("invalid alias"))
			return
		}

		resURL, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)
			render.JSON(w, r, response.Error("not found"))
			return
		}
		if err != nil {
			log.Info("failes to get url", sl.Err(err))
			render.JSON(w, r, response.Error("internal error"))
			return
		}
		log.Info("got url", slog.String("url", resURL))
		http.Redirect(w, r, resURL, http.StatusFound)
	}
}
