package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/Irongoshan-ux/url-shortener/internal/auth"
	urlhandler "github.com/Irongoshan-ux/url-shortener/internal/handler"
	"github.com/Irongoshan-ux/url-shortener/internal/repository"
	"github.com/Irongoshan-ux/url-shortener/internal/service"
)

// newExampleServer returns a test HTTP server with the same routing pattern as production
// (auth cookie middleware + handler routes) and an HTTP client that keeps cookies.
func newExampleServer() (baseURL string, client *http.Client, cleanup func()) {
	jar, _ := cookiejar.New(nil)
	client = &http.Client{Jar: jar}

	var seq int
	repo := repository.NewMemoryRepository()
	svc := service.NewService(repo, service.WithIDGenerator(func() (string, error) {
		seq++
		return fmt.Sprintf("ex%d", seq), nil
	}))
	log := zerolog.Nop()
	h := urlhandler.NewHandler(svc, "http://example.com", log, nil)

	r := chi.NewRouter()
	r.Use(auth.CookieMiddleware("example-secret"))
	r.Mount("/", h.Router())

	ts := httptest.NewServer(r)
	return ts.URL, client, ts.Close
}

func Example_shortenPlain() {
	baseURL, client, cleanup := newExampleServer()
	defer cleanup()

	resp, err := client.Post(baseURL+"/", "text/plain", strings.NewReader("https://example.org/page"))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode)
	fmt.Println(strings.TrimSpace(string(body)))

	// Output:
	// 201
	// http://example.com/ex1
}

func Example_shortenJSON() {
	baseURL, client, cleanup := newExampleServer()
	defer cleanup()

	resp, err := client.Post(baseURL+"/api/shorten", "application/json", strings.NewReader(`{"url":"https://example.org/json"}`))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var out struct {
		Result string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return
	}
	fmt.Println(resp.StatusCode)
	fmt.Println(out.Result)

	// Output:
	// 201
	// http://example.com/ex1
}

func Example_shortenBatch() {
	baseURL, client, cleanup := newExampleServer()
	defer cleanup()

	payload := `[{"correlation_id":"a","original_url":"https://example.org/a"},{"correlation_id":"b","original_url":"https://example.org/b"}]`
	resp, err := client.Post(baseURL+"/api/shorten/batch", "application/json", strings.NewReader(payload))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var out []struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return
	}
	fmt.Println(resp.StatusCode)
	for _, row := range out {
		fmt.Println(row.CorrelationID, row.ShortURL)
	}

	// Output:
	// 201
	// a http://example.com/ex1
	// b http://example.com/ex2
}

func Example_redirect() {
	baseURL, client, cleanup := newExampleServer()
	defer cleanup()

	resp, err := client.Post(baseURL+"/", "text/plain", strings.NewReader("https://example.org/target"))
	if err != nil {
		return
	}
	resp.Body.Close()

	noFollow := &http.Client{
		Jar: client.Jar,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp2, err := noFollow.Get(baseURL + "/ex1")
	if err != nil {
		return
	}
	defer resp2.Body.Close()

	fmt.Println(resp2.StatusCode)
	fmt.Println(resp2.Header.Get("Location"))

	// Output:
	// 307
	// https://example.org/target
}

func Example_userURLs() {
	baseURL, client, cleanup := newExampleServer()
	defer cleanup()

	resp, err := client.Post(baseURL+"/", "text/plain", strings.NewReader("https://example.org/mine"))
	if err != nil {
		return
	}
	resp.Body.Close()

	resp2, err := client.Get(baseURL + "/api/user/urls")
	if err != nil {
		return
	}
	defer resp2.Body.Close()

	body, err := io.ReadAll(resp2.Body)
	if err != nil {
		return
	}
	var items []map[string]any
	if err := json.Unmarshal(body, &items); err != nil {
		return
	}
	fmt.Println(resp2.StatusCode)
	fmt.Println(len(items))
	fmt.Println(items[0]["original_url"])

	// Output:
	// 200
	// 1
	// https://example.org/mine
}

func Example_deleteUserURLs() {
	baseURL, client, cleanup := newExampleServer()
	defer cleanup()

	resp, err := client.Post(baseURL+"/", "text/plain", strings.NewReader("https://example.org/del"))
	if err != nil {
		return
	}
	resp.Body.Close()

	req, err := http.NewRequest(http.MethodDelete, baseURL+"/api/user/urls", bytes.NewReader([]byte(`["ex1"]`)))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp2, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp2.Body.Close()

	fmt.Println(resp2.StatusCode)

	// Output:
	// 202
}
