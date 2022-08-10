package tests

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/bananocoin/boompow/apps/server/src/database"
	"github.com/bananocoin/boompow/apps/server/src/middleware"
	"github.com/bananocoin/boompow/apps/server/src/repository"
	"github.com/bananocoin/boompow/libs/utils/auth"
	utils "github.com/bananocoin/boompow/libs/utils/testing"
	"github.com/go-chi/chi"
)

func TestAuthMiddleware(t *testing.T) {
	os.Setenv("MOCK_REDIS", "true")
	mockDb, err := database.NewConnection(&database.Config{
		Host:     os.Getenv("DB_MOCK_HOST"),
		Port:     os.Getenv("DB_MOCK_PORT"),
		Password: os.Getenv("DB_MOCK_PASS"),
		User:     os.Getenv("DB_MOCK_USER"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		DBName:   "testing",
	})
	utils.AssertEqual(t, nil, err)
	err = database.DropAndCreateTables(mockDb)
	utils.AssertEqual(t, nil, err)
	userRepo := repository.NewUserService(mockDb)
	userRepo.CreateMockUsers()

	publicEndpoint := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("banano"))
	}

	// Endpoint that requires auth from the people that provide work
	authorizedProviderEndpoint := func(w http.ResponseWriter, r *http.Request) {
		// Only PROVIDER type users can provide work
		provider := middleware.AuthorizedProvider(r.Context())
		if provider == nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("401 - Unauthorized"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("banano"))
	}

	// Endpoint that requires auth from the people that request work
	authorizedRequesterEndpoint := func(w http.ResponseWriter, r *http.Request) {
		// Only REQUESTER type users can provide work
		requester := middleware.AuthorizedRequester(r.Context())
		if requester == nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("401 - Unauthorized"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("banano"))
	}

	// Endpoint that requires a token
	authorizedTokenEndpoint := func(w http.ResponseWriter, r *http.Request) {
		// Only tokens work
		requester := middleware.AuthorizedServiceToken(r.Context())
		if requester == nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("401 - Unauthorized"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("banano"))
	}

	authMiddleware := middleware.AuthMiddleware(userRepo)
	router := chi.NewRouter()
	router.Use(authMiddleware)
	router.Get("/", publicEndpoint)
	router.Get("/authProvider", authorizedProviderEndpoint)
	router.Get("/authRequester", authorizedRequesterEndpoint)
	router.Get("/authToken", authorizedTokenEndpoint)
	ts := httptest.NewServer(router)
	defer ts.Close()

	// Test that middleware doesn't block public endpoints
	if resp, body := testRequest(t, ts, "GET", "/", nil, ""); body != "banano" && resp.StatusCode != http.StatusOK {
		t.Fatalf(body)
	}

	// Test private jwt endpoint is blocked when not authorized
	if resp, body := testRequest(t, ts, "GET", "/authProvider", nil, ""); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf(body)
	}

	// Get a JWT token
	providerToken, _ := auth.GenerateToken("provider@gmail.com", time.Now)
	requesterToken, _ := auth.GenerateToken("requester@gmail.com", time.Now)

	// Endpoint that is only good for providers

	// Test private jwt endpoint is not blocked when authorized
	if resp, body := testRequest(t, ts, "GET", "/authProvider", nil, providerToken); resp.StatusCode != http.StatusOK && body != "banano" {
		t.Fatalf(body)
	}
	// Test private jwt endpoint is blocked for requester
	if resp, body := testRequest(t, ts, "GET", "/authProvider", nil, requesterToken); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf(body)
	}

	// Endpoint that is only good for requesters

	// Test private jwt endpoint is not blocked when authorized
	if resp, body := testRequest(t, ts, "GET", "/authRequester", nil, requesterToken); resp.StatusCode != http.StatusOK && body != "banano" {
		t.Fatalf(body)
	}
	// Test private jwt endpoint is blocked for provider
	if resp, body := testRequest(t, ts, "GET", "/authRequester", nil, providerToken); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf(body)
	}

	// Endpoint that is only good for tokens (not jwt)
	if resp, body := testRequest(t, ts, "GET", "/authToken", nil, providerToken); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf(body)
	}
	if resp, body := testRequest(t, ts, "GET", "/authToken", nil, requesterToken); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf(body)
	}
	// Create the token
	// Generate token
	token := userRepo.GenerateServiceToken()
	requesterEmail := "requester@gmail.com"
	// Get user
	requester, _ := userRepo.GetUser(nil, &requesterEmail)
	if err := database.GetRedisDB().AddServiceToken(requester.ID, token); err != nil {
		t.Errorf("Error adding service token to redis: %s", err.Error())
	}

	if resp, body := testRequest(t, ts, "GET", "/authToken", nil, token); resp.StatusCode != http.StatusOK && body != "banano" {
		t.Fatalf(body)
	}
	// Test a random token
	if resp, body := testRequest(t, ts, "GET", "/authToken", nil, userRepo.GenerateServiceToken()); resp.StatusCode != http.StatusForbidden {
		t.Fatalf(body)
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader, token string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}
	defer resp.Body.Close()

	return resp, string(respBody)
}
