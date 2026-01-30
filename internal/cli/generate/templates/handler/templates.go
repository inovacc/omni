package handler

// TemplateData contains all data needed for handler template rendering
type TemplateData struct {
	Name       string   // Handler name (e.g., "User")
	NameLower  string   // Lowercase name (e.g., "user")
	Package    string   // Package name
	Methods    []string // HTTP methods (GET, POST, PUT, DELETE, PATCH)
	Path       string   // URL path pattern
	Middleware bool     // Include middleware support
	Framework  string   // Framework: stdlib, chi, gin, echo
}

// StdlibHandlerTemplate generates a net/http handler
const StdlibHandlerTemplate = `package {{.Package}}

import (
	"encoding/json"
	"net/http"
)

// {{.Name}}Handler handles {{.NameLower}} operations
type {{.Name}}Handler struct {
	// Add dependencies here (e.g., service, repository)
}

// New{{.Name}}Handler creates a new {{.Name}}Handler
func New{{.Name}}Handler() *{{.Name}}Handler {
	return &{{.Name}}Handler{}
}
{{range .Methods}}
{{if eq . "GET"}}
// Get handles GET requests
func (h *{{$.Name}}Handler) Get(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement GET logic
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "GET {{$.NameLower}}"})
}
{{end}}
{{if eq . "POST"}}
// Create handles POST requests
func (h *{{$.Name}}Handler) Create(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement POST logic
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "POST {{$.NameLower}}"})
}
{{end}}
{{if eq . "PUT"}}
// Update handles PUT requests
func (h *{{$.Name}}Handler) Update(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement PUT logic
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "PUT {{$.NameLower}}"})
}
{{end}}
{{if eq . "DELETE"}}
// Delete handles DELETE requests
func (h *{{$.Name}}Handler) Delete(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement DELETE logic
	w.WriteHeader(http.StatusNoContent)
}
{{end}}
{{if eq . "PATCH"}}
// Patch handles PATCH requests
func (h *{{$.Name}}Handler) Patch(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement PATCH logic
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "PATCH {{$.NameLower}}"})
}
{{end}}
{{end}}
// ServeHTTP implements http.Handler
func (h *{{.Name}}Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
{{- range .Methods}}
	case http.Method{{.}}:
		{{if eq . "GET"}}h.Get(w, r){{end}}
		{{- if eq . "POST"}}h.Create(w, r){{end}}
		{{- if eq . "PUT"}}h.Update(w, r){{end}}
		{{- if eq . "DELETE"}}h.Delete(w, r){{end}}
		{{- if eq . "PATCH"}}h.Patch(w, r){{end}}
{{- end}}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
`

// ChiHandlerTemplate generates a chi router handler
const ChiHandlerTemplate = `package {{.Package}}

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// {{.Name}}Handler handles {{.NameLower}} operations
type {{.Name}}Handler struct {
	// Add dependencies here (e.g., service, repository)
}

// New{{.Name}}Handler creates a new {{.Name}}Handler
func New{{.Name}}Handler() *{{.Name}}Handler {
	return &{{.Name}}Handler{}
}

// Routes returns a chi.Router with all routes registered
func (h *{{.Name}}Handler) Routes() chi.Router {
	r := chi.NewRouter()
{{if .Middleware}}
	// Add middleware here
	// r.Use(middleware.Logger)
{{end}}
{{range .Methods}}
{{if eq . "GET"}}
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
{{end}}
{{- if eq . "POST"}}
	r.Post("/", h.Create)
{{end}}
{{- if eq . "PUT"}}
	r.Put("/{id}", h.Update)
{{end}}
{{- if eq . "DELETE"}}
	r.Delete("/{id}", h.Delete)
{{end}}
{{- if eq . "PATCH"}}
	r.Patch("/{id}", h.Patch)
{{end}}
{{- end}}

	return r
}
{{range .Methods}}
{{if eq . "GET"}}
// List handles GET / requests
func (h *{{$.Name}}Handler) List(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement list logic
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode([]map[string]string{{"message": "list {{$.NameLower}}"}})
}

// Get handles GET /{id} requests
func (h *{{$.Name}}Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// TODO: Implement get logic
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"id": id, "message": "get {{$.NameLower}}"})
}
{{end}}
{{if eq . "POST"}}
// Create handles POST / requests
func (h *{{$.Name}}Handler) Create(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement create logic
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "created {{$.NameLower}}"})
}
{{end}}
{{if eq . "PUT"}}
// Update handles PUT /{id} requests
func (h *{{$.Name}}Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// TODO: Implement update logic
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"id": id, "message": "updated {{$.NameLower}}"})
}
{{end}}
{{if eq . "DELETE"}}
// Delete handles DELETE /{id} requests
func (h *{{$.Name}}Handler) Delete(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "id")
	// TODO: Implement delete logic
	w.WriteHeader(http.StatusNoContent)
}
{{end}}
{{if eq . "PATCH"}}
// Patch handles PATCH /{id} requests
func (h *{{$.Name}}Handler) Patch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// TODO: Implement patch logic
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"id": id, "message": "patched {{$.NameLower}}"})
}
{{end}}
{{end}}
`

// GinHandlerTemplate generates a gin handler
const GinHandlerTemplate = `package {{.Package}}

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// {{.Name}}Handler handles {{.NameLower}} operations
type {{.Name}}Handler struct {
	// Add dependencies here (e.g., service, repository)
}

// New{{.Name}}Handler creates a new {{.Name}}Handler
func New{{.Name}}Handler() *{{.Name}}Handler {
	return &{{.Name}}Handler{}
}

// RegisterRoutes registers all routes for this handler
func (h *{{.Name}}Handler) RegisterRoutes(r *gin.RouterGroup) {
{{if .Middleware}}
	// Add middleware here
	// r.Use(gin.Logger())
{{end}}
{{range .Methods}}
{{if eq . "GET"}}
	r.GET("", h.List)
	r.GET("/:id", h.Get)
{{end}}
{{- if eq . "POST"}}
	r.POST("", h.Create)
{{end}}
{{- if eq . "PUT"}}
	r.PUT("/:id", h.Update)
{{end}}
{{- if eq . "DELETE"}}
	r.DELETE("/:id", h.Delete)
{{end}}
{{- if eq . "PATCH"}}
	r.PATCH("/:id", h.Patch)
{{end}}
{{- end}}
}
{{range .Methods}}
{{if eq . "GET"}}
// List handles GET requests
func (h *{{$.Name}}Handler) List(c *gin.Context) {
	// TODO: Implement list logic
	c.JSON(http.StatusOK, []gin.H{{"message": "list {{$.NameLower}}"}})
}

// Get handles GET /:id requests
func (h *{{$.Name}}Handler) Get(c *gin.Context) {
	id := c.Param("id")
	// TODO: Implement get logic
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "get {{$.NameLower}}"})
}
{{end}}
{{if eq . "POST"}}
// Create handles POST requests
func (h *{{$.Name}}Handler) Create(c *gin.Context) {
	// TODO: Implement create logic
	c.JSON(http.StatusCreated, gin.H{"message": "created {{$.NameLower}}"})
}
{{end}}
{{if eq . "PUT"}}
// Update handles PUT /:id requests
func (h *{{$.Name}}Handler) Update(c *gin.Context) {
	id := c.Param("id")
	// TODO: Implement update logic
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "updated {{$.NameLower}}"})
}
{{end}}
{{if eq . "DELETE"}}
// Delete handles DELETE /:id requests
func (h *{{$.Name}}Handler) Delete(c *gin.Context) {
	_ = c.Param("id")
	// TODO: Implement delete logic
	c.Status(http.StatusNoContent)
}
{{end}}
{{if eq . "PATCH"}}
// Patch handles PATCH /:id requests
func (h *{{$.Name}}Handler) Patch(c *gin.Context) {
	id := c.Param("id")
	// TODO: Implement patch logic
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "patched {{$.NameLower}}"})
}
{{end}}
{{end}}
`

// EchoHandlerTemplate generates an echo handler
const EchoHandlerTemplate = `package {{.Package}}

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// {{.Name}}Handler handles {{.NameLower}} operations
type {{.Name}}Handler struct {
	// Add dependencies here (e.g., service, repository)
}

// New{{.Name}}Handler creates a new {{.Name}}Handler
func New{{.Name}}Handler() *{{.Name}}Handler {
	return &{{.Name}}Handler{}
}

// RegisterRoutes registers all routes for this handler
func (h *{{.Name}}Handler) RegisterRoutes(g *echo.Group) {
{{if .Middleware}}
	// Add middleware here
	// g.Use(middleware.Logger())
{{end}}
{{range .Methods}}
{{if eq . "GET"}}
	g.GET("", h.List)
	g.GET("/:id", h.Get)
{{end}}
{{- if eq . "POST"}}
	g.POST("", h.Create)
{{end}}
{{- if eq . "PUT"}}
	g.PUT("/:id", h.Update)
{{end}}
{{- if eq . "DELETE"}}
	g.DELETE("/:id", h.Delete)
{{end}}
{{- if eq . "PATCH"}}
	g.PATCH("/:id", h.Patch)
{{end}}
{{- end}}
}
{{range .Methods}}
{{if eq . "GET"}}
// List handles GET requests
func (h *{{$.Name}}Handler) List(c echo.Context) error {
	// TODO: Implement list logic
	return c.JSON(http.StatusOK, []map[string]string{{"message": "list {{$.NameLower}}"}})
}

// Get handles GET /:id requests
func (h *{{$.Name}}Handler) Get(c echo.Context) error {
	id := c.Param("id")
	// TODO: Implement get logic
	return c.JSON(http.StatusOK, map[string]string{"id": id, "message": "get {{$.NameLower}}"})
}
{{end}}
{{if eq . "POST"}}
// Create handles POST requests
func (h *{{$.Name}}Handler) Create(c echo.Context) error {
	// TODO: Implement create logic
	return c.JSON(http.StatusCreated, map[string]string{"message": "created {{$.NameLower}}"})
}
{{end}}
{{if eq . "PUT"}}
// Update handles PUT /:id requests
func (h *{{$.Name}}Handler) Update(c echo.Context) error {
	id := c.Param("id")
	// TODO: Implement update logic
	return c.JSON(http.StatusOK, map[string]string{"id": id, "message": "updated {{$.NameLower}}"})
}
{{end}}
{{if eq . "DELETE"}}
// Delete handles DELETE /:id requests
func (h *{{$.Name}}Handler) Delete(c echo.Context) error {
	_ = c.Param("id")
	// TODO: Implement delete logic
	return c.NoContent(http.StatusNoContent)
}
{{end}}
{{if eq . "PATCH"}}
// Patch handles PATCH /:id requests
func (h *{{$.Name}}Handler) Patch(c echo.Context) error {
	id := c.Param("id")
	// TODO: Implement patch logic
	return c.JSON(http.StatusOK, map[string]string{"id": id, "message": "patched {{$.NameLower}}"})
}
{{end}}
{{end}}
`

// HandlerTestTemplate generates handler tests
const HandlerTestTemplate = `package {{.Package}}

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew{{.Name}}Handler(t *testing.T) {
	h := New{{.Name}}Handler()
	if h == nil {
		t.Fatal("New{{.Name}}Handler() returned nil")
	}
}
{{range .Methods}}
{{if eq . "GET"}}
func Test{{$.Name}}Handler_Get(t *testing.T) {
	h := New{{$.Name}}Handler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.Get(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Get() status = %d, want %d", w.Code, http.StatusOK)
	}
}
{{end}}
{{if eq . "POST"}}
func Test{{$.Name}}Handler_Create(t *testing.T) {
	h := New{{$.Name}}Handler()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Create() status = %d, want %d", w.Code, http.StatusCreated)
	}
}
{{end}}
{{if eq . "DELETE"}}
func Test{{$.Name}}Handler_Delete(t *testing.T) {
	h := New{{$.Name}}Handler()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	w := httptest.NewRecorder()

	h.Delete(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Delete() status = %d, want %d", w.Code, http.StatusNoContent)
	}
}
{{end}}
{{end}}
`
