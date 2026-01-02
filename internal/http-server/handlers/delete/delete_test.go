package delete_test

import (
	// project
	"go-url-shortener/internal/http-server/handlers/delete"
	"go-url-shortener/internal/http-server/handlers/delete/mocks"
	"go-url-shortener/internal/lib/logger/handlers/slogdiscard"
	"go-url-shortener/internal/storage"
	
	// embedded
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	// external
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

)

func TestDeleteHandler(t *testing.T) {
	cases := []struct {
		name       string
		alias      string
		mockError  error
		statusCode int
		respError  string
	}{
		{
			name:       "Success",
			alias:      "test_alias",
			statusCode: http.StatusOK,
		},
		{
			name:       "URL Not Found",
			alias:      "missing",
			mockError:  storage.ErrURLNotFound,
			statusCode: http.StatusOK,
			respError:  "not found",
		},
		{
			name:       "Internal Error",
			alias:      "broken",
			mockError:  errors.New("db down"),
			statusCode: http.StatusOK,
			respError:  "internal error",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deleter := mocks.NewURLDeleter(t)

			deleter.
				On("DeleteURL", tc.alias).
				Return(tc.mockError).
				Once()

			r := chi.NewRouter()
			r.Delete("/{alias}", delete.New(
				slogdiscard.NewDiscardLogger(),
				deleter,
			))

			req := httptest.NewRequest(
				http.MethodDelete,
				"/"+tc.alias,
				nil,
			)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.statusCode, rec.Code)

			if tc.respError != "" {
				var resp map[string]string
				err := json.NewDecoder(rec.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, tc.respError, resp["error"])
			}
		})
	}
}
