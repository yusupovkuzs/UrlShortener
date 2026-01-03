package tests

import (
	// project
	"go-url-shortener/internal/config"
	"go-url-shortener/internal/http-server/handlers/delete"
	"go-url-shortener/internal/http-server/handlers/redirect"
	"go-url-shortener/internal/http-server/handlers/save"
	"go-url-shortener/internal/http-server/middleware/logger"
	"go-url-shortener/internal/lib/logger/handlers/slogpretty"
	"go-url-shortener/internal/storage/postgres"

	// embedded
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	// external
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func TestURLShortener_FullCRUD(t *testing.T) {
	cfg := config.MustLoad()

	log := setupPrettySlog()

	st, err := postgres.NewStorage(cfg.Postgres)
	if err != nil {
		t.Fatalf("postgres init failed: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(logger.New(log))
	r.Use(middleware.Recoverer)

	r.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth(
			"url-shortener",
			map[string]string{
				cfg.HttpServer.User: cfg.HttpServer.Password,
			},
		))
		r.Post("/", save.New(log, st))
		r.Delete("/{alias}", delete.New(log, st))
	})

	r.Get("/{alias}", redirect.New(log, st))

	server := httptest.NewServer(r)
	defer server.Close()

	auth := "Basic " + base64.StdEncoding.EncodeToString(
		[]byte(cfg.HttpServer.User+":"+cfg.HttpServer.Password),
	)

	saveBody := map[string]string{
		"url": "https://example.com/test",
	}

	body, _ := json.Marshal(saveBody)

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/url/", bytes.NewReader(body))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("save request failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	var saveResp struct {
		Alias string `json:"alias"`
	}
	_ = json.NewDecoder(res.Body).Decode(&saveResp)

	if saveResp.Alias == "" {
		t.Fatal("alias is empty")
	}

	alias := saveResp.Alias

	req, _ = http.NewRequest(http.MethodGet, server.URL+"/"+alias, nil)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	res, err = client.Do(req)

	if err != nil {
		t.Fatalf("redirect request failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusFound {
		t.Fatalf("expected 302, got %d", res.StatusCode)
	}

	if loc := res.Header.Get("Location"); loc != "https://example.com/test" {
		t.Fatalf("unexpected redirect location: %s", loc)
	}

	req, _ = http.NewRequest(http.MethodDelete, server.URL+"/url/"+alias, nil)
	req.Header.Set("Authorization", auth)

	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("delete request failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	req, _ = http.NewRequest(http.MethodGet, server.URL+"/"+alias, nil)
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get after delete failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with error body, got %d", res.StatusCode)
	}

	var errResp map[string]string
	_ = json.NewDecoder(res.Body).Decode(&errResp)

	if errResp["error"] != "not found" {
		t.Fatalf("expected not found, got %v", errResp)
	}
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
