package azure

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetAADAccessTokenViaClientAssertion_Success(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "application/x-www-form-urlencoded") {
			t.Fatalf("expected form content-type, got %s", ct)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed parsing form: %v", err)
		}
		assertEq(t, r.Form.Get("client_id"), "client")
		assertEq(t, r.Form.Get("grant_type"), "client_credentials")
		assertEq(t, r.Form.Get("client_assertion_type"), "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
		assertEq(t, r.Form.Get("client_assertion"), "idtoken")
		assertEq(t, r.Form.Get("scope"), DefaultResource+".default")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token":"AT","token_type":"Bearer","expires_in":3600}`))
	}))
	defer ts.Close()

	tok, err := GetAADAccessTokenViaClientAssertion(context.Background(), "tenant", "client", "idtoken", ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "AT" {
		t.Fatalf("expected access token AT, got %q", tok)
	}
}

func TestGetAADAccessTokenViaClientAssertion_400WithErrorField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_client","error_description":"bad"}`))
	}))
	defer ts.Close()

	_, err := GetAADAccessTokenViaClientAssertion(context.Background(), "tenant", "client", "idtoken", ts.URL)
	if err == nil || !strings.Contains(err.Error(), "status=400") || !strings.Contains(err.Error(), "invalid_client") {
		t.Fatalf("expected 400 with invalid_client error, got %v", err)
	}
}

func TestGetAADAccessTokenViaClientAssertion_400WithoutErrorField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("{}"))
	}))
	defer ts.Close()

	_, err := GetAADAccessTokenViaClientAssertion(context.Background(), "tenant", "client", "idtoken", ts.URL)
	if err == nil || !strings.Contains(err.Error(), "status=400") {
		t.Fatalf("expected 400 error, got %v", err)
	}
}

func TestGetAADAccessTokenViaClientAssertion_MalformedJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json"))
	}))
	defer ts.Close()

	_, err := GetAADAccessTokenViaClientAssertion(context.Background(), "tenant", "client", "idtoken", ts.URL)
	if err == nil {
		t.Fatalf("expected JSON decode error, got nil")
	}
}

func TestGetAADAccessTokenViaClientAssertion_MissingAccessToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"token_type":"Bearer","expires_in":3600}`))
	}))
	defer ts.Close()

	_, err := GetAADAccessTokenViaClientAssertion(context.Background(), "tenant", "client", "idtoken", ts.URL)
	if err == nil || !strings.Contains(err.Error(), "missing access_token") {
		t.Fatalf("expected missing access_token error, got %v", err)
	}
}

func assertEq(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("mismatch: got=%q want=%q", got, want)
	}
}

