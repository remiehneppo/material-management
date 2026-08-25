package session

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRefreshCookieAttributes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name, host         string
		production, secure bool
	}{
		{"local development", "localhost:3000", false, false},
		{"remote development", "dev.example.com", false, true},
		{"production", "app.example.com", true, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest("POST", "http://"+test.host+"/api/v1/auth/login", nil)
			NewHandler(nil, test.production).setRefreshCookie(ctx, "opaque")
			cookie := recorder.Header().Get("Set-Cookie")
			for _, required := range []string{"HttpOnly", "SameSite=Lax", "Path=/api/v1/auth"} {
				if !strings.Contains(cookie, required) {
					t.Fatalf("cookie %q missing %s", cookie, required)
				}
			}
			if strings.Contains(cookie, "Secure") != test.secure {
				t.Fatalf("cookie %q secure=%v", cookie, test.secure)
			}
		})
	}
}
