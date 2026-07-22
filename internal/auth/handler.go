package auth

import (
	"html/template"
	"net/http"
)

type CookieManager interface {
	SetSession(http.ResponseWriter, string)
	ClearSession(http.ResponseWriter)
}

type Handler struct {
	Authenticator Authenticator
	Templates     *template.Template
	Cookies       CookieManager
}

func (h Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	_ = h.Templates.ExecuteTemplate(w, "login.html", map[string]any{"Title": "Login"})
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	if err := h.Authenticator.Authenticate(username, password); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	h.Cookies.SetSession(w, username)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h Handler) Logout(w http.ResponseWriter, r *http.Request) {
	h.Cookies.ClearSession(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
