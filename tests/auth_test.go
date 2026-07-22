package tests

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"voidfs/internal/config"
	"voidfs/internal/server"
)

type fakeAuthenticator struct {
	username string
	password string
}

func (f fakeAuthenticator) Authenticate(username, password string) error {
	if username == f.username && password == f.password {
		return nil
	}
	return errors.New("invalid credentials")
}

func newAuthTestApp(t *testing.T) http.Handler {
	t.Helper()
	cfg := config.Load()
	cfg.RootDir = t.TempDir()
	cfg.SessionSecret = "test-secret"
	return server.NewWithAuthenticator(cfg, fakeAuthenticator{
		username: "vps-user",
		password: "vps-password",
	}).Router()
}

func TestProtectedRouteRedirectsToLoginWithoutSession(t *testing.T) {
	app := newAuthTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/files?path=/", nil)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected %d, got %d", http.StatusSeeOther, rr.Code)
	}
	if location := rr.Header().Get("Location"); location != "/login" {
		t.Fatalf("expected redirect to /login, got %q", location)
	}
}

func TestLoginUsesVPSAuthenticatorAndSetsSession(t *testing.T) {
	app := newAuthTestApp(t)
	form := url.Values{}
	form.Set("username", "vps-user")
	form.Set("password", "vps-password")

	loginReq := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRR := httptest.NewRecorder()
	app.ServeHTTP(loginRR, loginReq)

	if loginRR.Code != http.StatusSeeOther {
		t.Fatalf("expected %d, got %d body=%s", http.StatusSeeOther, loginRR.Code, loginRR.Body.String())
	}
	cookies := loginRR.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie, got none")
	}

	protectedReq := httptest.NewRequest(http.MethodGet, "/files?path=/", nil)
	protectedReq.AddCookie(cookies[0])
	protectedRR := httptest.NewRecorder()
	app.ServeHTTP(protectedRR, protectedReq)
	if protectedRR.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d body=%s", http.StatusOK, protectedRR.Code, protectedRR.Body.String())
	}
}

func TestLoginRejectsInvalidVPSCredentials(t *testing.T) {
	app := newAuthTestApp(t)
	form := url.Values{}
	form.Set("username", "vps-user")
	form.Set("password", "wrong")

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if len(rr.Result().Cookies()) != 0 {
		t.Fatal("expected no session cookie for invalid credentials")
	}
}

func TestLogoutClearsSession(t *testing.T) {
	app := newAuthTestApp(t)
	form := url.Values{}
	form.Set("username", "vps-user")
	form.Set("password", "vps-password")

	loginReq := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRR := httptest.NewRecorder()
	app.ServeHTTP(loginRR, loginReq)
	cookies := loginRR.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie, got none")
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/logout", nil)
	logoutReq.AddCookie(cookies[0])
	logoutRR := httptest.NewRecorder()
	app.ServeHTTP(logoutRR, logoutReq)

	if logoutRR.Code != http.StatusSeeOther {
		t.Fatalf("expected %d, got %d", http.StatusSeeOther, logoutRR.Code)
	}
	setCookie := logoutRR.Header().Get("Set-Cookie")
	if !strings.Contains(setCookie, "Max-Age=0") {
		t.Fatalf("expected cleared cookie, got %q", setCookie)
	}
}
