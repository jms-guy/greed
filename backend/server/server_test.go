package server_test

/*
import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/jms-guy/greed/backend/server/handlers"
	"github.com/jms-guy/greed/models"
)

// End-to-End server integration test
func TestE2E(t *testing.T) {
	app, err := handlers.NewAppServer()
	if err != nil {
		t.Fatal(err)
	}

	handler := app.Router()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	e := httpexpect.Default(t, ts.URL)

	// Ping test
	e.GET("/api/health").Expect().Status(http.StatusOK).Text()

	// Register
	e.POST("/api/auth/register").
		WithJSON(models.UserDetails{Name: "Test", Password: "password", Email: "test@email.com"}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().
		HasValue("name", "Test")

	// Login
	loginResp := e.POST("/api/auth/login").
		WithJSON(models.UserDetails{Name: "Test", Password: "password", Email: "test@email.com"}).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	loginResp.HasValue("token_type", "Bearer")
	loginResp.ContainsKey("access_token")
	loginResp.ContainsKey("refresh_token")

	jwtToken := loginResp.Value("access_token").String().Raw()
	refreshToken := loginResp.Value("refresh_token").String().Raw()

	// Item creation
	e.POST("/admin/sandbox").
		WithHeader("Authorization", "Bearer "+jwtToken).
		Expect().
		Status(http.StatusCreated)

	// Item get
	userItems := e.GET("/api/items").
		WithHeader("Authorization", "Bearer "+jwtToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	itemsArray := userItems.Value("items").Array()
	itemsArray.Length().Gt(0)

	// Refresh JWT
	newJWT := e.POST("/api/auth/refresh").
		WithJSON(models.RefreshRequest{RefreshToken: refreshToken}).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	newJWT.HasValue("token_type", "Bearer")
	newJWT.ContainsKey("refresh_token")
	newJWT.ContainsKey("access_token")

	newRefreshToken := newJWT.Value("refresh_token").String().Raw()

	// Logout
	e.POST("/api/auth/logout").
		WithJSON(models.RefreshRequest{RefreshToken: newRefreshToken}).
		Expect().
		Status(http.StatusOK)
}
*/
