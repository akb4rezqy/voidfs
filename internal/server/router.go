package server

import (
	"html/template"
	"net/http"
	"path/filepath"

	authsvc "voidfs/internal/auth"
	"voidfs/internal/config"
	editorsvc "voidfs/internal/editor"
	filesvc "voidfs/internal/files"
	uploadsvc "voidfs/internal/upload"
)

type Server struct {
	cfg           config.Config
	templates     *template.Template
	authenticator authsvc.Authenticator
}

type cookieManager struct {
	secret string
}

func (c cookieManager) SetSession(w http.ResponseWriter, username string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    signSessionValue(c.secret, username),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (c cookieManager) ClearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func loadTemplates() *template.Template {
	patterns := []string{
		"internal/views/*.html",
		"internal/views/partials/*.html",
		"internal/views/components/*.html",
		filepath.Join("..", "internal", "views", "*.html"),
		filepath.Join("..", "internal", "views", "partials", "*.html"),
		filepath.Join("..", "internal", "views", "components", "*.html"),
	}
	return template.Must(template.ParseGlob(patterns[0]))
}

func mustParseTemplates() *template.Template {
	basePatterns := []string{
		"internal/views/*.html",
		"internal/views/partials/*.html",
		"internal/views/components/*.html",
	}
	if tmpl, err := template.ParseGlob(basePatterns[0]); err == nil {
		_, _ = tmpl.ParseGlob(basePatterns[1])
		_, _ = tmpl.ParseGlob(basePatterns[2])
		return tmpl
	}
	fallbackPatterns := []string{
		filepath.Join("..", "internal", "views", "*.html"),
		filepath.Join("..", "internal", "views", "partials", "*.html"),
		filepath.Join("..", "internal", "views", "components", "*.html"),
	}
	tmpl := template.Must(template.ParseGlob(fallbackPatterns[0]))
	_, _ = tmpl.ParseGlob(fallbackPatterns[1])
	_, _ = tmpl.ParseGlob(fallbackPatterns[2])
	return tmpl
}

func New(cfg config.Config) *Server {
	return NewWithAuthenticator(cfg, authsvc.NewPAMAuthenticator("voidfs", cfg.AllowedUser))
}

func NewWithAuthenticator(cfg config.Config, authenticator authsvc.Authenticator) *Server {
	return &Server{
		cfg:           cfg,
		templates:     mustParseTemplates(),
		authenticator: authenticator,
	}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()
	fileService := filesvc.NewService(s.cfg.RootDir)
	editorService := editorsvc.NewService(s.cfg.RootDir, s.cfg.MaxEditBytes)
	uploadService := uploadsvc.NewService(s.cfg.RootDir, s.cfg.MaxUploadBytes)
	fileHandler := filesvc.Handler{Service: fileService}
	editorHandler := editorsvc.Handler{Service: editorService}
	uploadHandler := uploadsvc.Handler{Service: uploadService}
	authHandler := authsvc.Handler{
		Authenticator: s.authenticator,
		Templates:     s.templates,
		Cookies:       cookieManager{secret: s.cfg.SessionSecret},
	}

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			authHandler.LoginPage(w, r)
			return
		}
		if r.Method == http.MethodPost {
			authHandler.Login(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/logout", authHandler.Logout)
	mux.HandleFunc("/", s.requireAuth(s.handleHome))
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/files", s.requireAuth(fileHandler.List))
	mux.HandleFunc("/files/folder", s.requireAuth(fileHandler.CreateFolder))
	mux.HandleFunc("/files/file", s.requireAuth(fileHandler.CreateFile))
	mux.HandleFunc("/files/rename", s.requireAuth(fileHandler.Rename))
	mux.HandleFunc("/files/delete", s.requireAuth(fileHandler.Delete))
	mux.HandleFunc("/files/delete-many", s.requireAuth(fileHandler.DeleteMany))
	mux.HandleFunc("/files/zip", s.requireAuth(fileHandler.CreateZip))
	mux.HandleFunc("/files/unzip", s.requireAuth(fileHandler.ExtractZips))
	mux.HandleFunc("/files/download", s.requireAuth(fileHandler.Download))
	mux.HandleFunc("/editor/open", s.requireAuth(editorHandler.Open))
	mux.HandleFunc("/editor/save", s.requireAuth(editorHandler.Save))
	mux.HandleFunc("/upload", s.requireAuth(uploadHandler.Upload))
	staticFiles := http.StripPrefix("/static/", http.FileServer(http.Dir("web/static")))
	mux.Handle("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		staticFiles.ServeHTTP(w, r)
	}))
	return mux
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	data := map[string]any{
		"Title":   "VoidFS",
		"RootDir": s.cfg.RootDir,
	}

	if err := s.templates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok"))
}
