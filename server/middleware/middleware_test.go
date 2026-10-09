package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestLimiter(t *testing.T) {
	l := NewLimiter(2, time.Minute)
	if !l.Allow("k") || !l.Allow("k") {
		t.Fatal("first two hits should pass")
	}
	if l.Allow("k") {
		t.Fatal("third hit should be limited")
	}
	if !l.Allow("other") {
		t.Fatal("keys must be independent")
	}
	l.Reset("k")
	if !l.Allow("k") {
		t.Fatal("reset should clear the key")
	}
}

func TestRequireJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequireJSON())
	r.POST("/x", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	cases := []struct {
		method, ctype string
		want          int
	}{
		{http.MethodPost, "application/json", http.StatusOK},
		{http.MethodPost, "application/json; charset=utf-8", http.StatusOK},
		{http.MethodPost, "text/plain", http.StatusUnsupportedMediaType},
		{http.MethodPost, "application/x-www-form-urlencoded", http.StatusUnsupportedMediaType},
		{http.MethodPost, "", http.StatusUnsupportedMediaType},
		{http.MethodGet, "", http.StatusOK},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, "/x", strings.NewReader(`{}`))
		if tc.ctype != "" {
			req.Header.Set("Content-Type", tc.ctype)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Errorf("%s %q: want %d, got %d", tc.method, tc.ctype, tc.want, w.Code)
		}
	}
}
