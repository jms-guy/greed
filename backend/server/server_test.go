package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/jms-guy/greed/backend/server/handlers"
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

	e.GET("/api/health").Expect().Status(http.StatusOK).Text()
}
