package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/gin-gonic/gin"
)

func TestSwaggerDocumentValid(t *testing.T) {
	spec, err := os.ReadFile(swaggerSpecPath(t))
	if err != nil {
		t.Fatal(err)
	}

	var document openapi2.T
	if err := json.Unmarshal(spec, &document); err != nil {
		t.Fatalf("parse Swagger document: %v", err)
	}
	if document.Swagger != "2.0" {
		t.Fatalf("Swagger version = %q, want 2.0", document.Swagger)
	}

	openAPI, err := openapi2conv.ToV3(&document)
	if err != nil {
		t.Fatalf("convert Swagger document: %v", err)
	}
	if err := openAPI.Validate(context.Background()); err != nil {
		t.Fatalf("validate Swagger document: %v", err)
	}
}

func TestSwaggerDocumentsAllAPIRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router, &Handler{}, nil)

	spec, err := os.ReadFile(swaggerSpecPath(t))
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(spec, &document); err != nil {
		t.Fatalf("parse Swagger document: %v", err)
	}

	documented := make(map[string]bool)
	for path, operations := range document.Paths {
		if !strings.HasPrefix(path, "/api/") {
			continue
		}
		for method := range operations {
			documented[strings.ToUpper(method)+" "+path] = true
		}
	}

	registered := make(map[string]bool)
	for _, route := range router.Routes() {
		if !strings.HasPrefix(route.Path, "/api/") {
			continue
		}
		key := route.Method + " " + swaggerPath(route.Path)
		registered[key] = true
		if !documented[key] {
			t.Errorf("route %s is not documented", key)
		}
	}
	for key := range documented {
		if !registered[key] {
			t.Errorf("documented operation %s is not registered", key)
		}
	}
}

func TestSwaggerDocumentsOperationalRoutes(t *testing.T) {
	spec, err := os.ReadFile(swaggerSpecPath(t))
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(spec, &document); err != nil {
		t.Fatalf("parse Swagger document: %v", err)
	}

	for _, key := range []string{"GET /health/live", "GET /health/ready"} {
		parts := strings.SplitN(key, " ", 2)
		if _, ok := document.Paths[parts[1]][strings.ToLower(parts[0])]; !ok {
			t.Errorf("operational route %s is not documented", key)
		}
	}
	if _, ok := document.Paths["/metrics"]; ok {
		t.Fatal("Prometheus metrics endpoint should not be included in Swagger")
	}
}

func TestRegisterSwaggerRoutes(t *testing.T) {
	tests := []struct {
		name        string
		enabled     bool
		environment string
		want        bool
		wantStatus  int
	}{
		{name: "enabled in development", enabled: true, environment: "development", want: true, wantStatus: http.StatusOK},
		{name: "disabled by flag", enabled: false, environment: "development", want: false, wantStatus: http.StatusNotFound},
		{name: "disabled in production", enabled: true, environment: "production", want: false, wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			RegisterSwaggerRoutes(router, tt.enabled, tt.environment)

			found := false
			for _, route := range router.Routes() {
				if route.Path == "/swagger/*any" {
					found = true
				}
			}
			if found != tt.want {
				t.Fatalf("Swagger route registered = %v, want %v", found, tt.want)
			}

			request := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != tt.wantStatus {
				t.Fatalf("Swagger UI status = %d, want %d", response.Code, tt.wantStatus)
			}
		})
	}
}

func swaggerSpecPath(t *testing.T) string {
	t.Helper()
	return filepath.Join("..", "..", "..", "..", "docs", "swagger.json")
}

func swaggerPath(path string) string {
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			segments[i] = "{" + strings.TrimPrefix(segment, ":") + "}"
		}
	}
	return strings.Join(segments, "/")
}
