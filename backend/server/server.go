package server

import (
	"net/http"
	"time"

	"github.com/jms-guy/greed/backend/server/handlers"
	_ "github.com/lib/pq"
)

func Run() error {
	// Create new AppServer struct
	app, err := handlers.NewAppServer()
	if err != nil {
		return err
	}

	// Initialize server router
	r := app.Router()

	server := &http.Server{
		Addr:         ":" + app.Config.Port,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	/////Start server/////
	_ = app.Logger.Log(
		"transport", "HTTP",
		"address", app.Config.Port,
		"msg", "listening",
	)

	err = server.ListenAndServe()
	if err != nil {
		_ = app.Logger.Log(
			"level", "error",
			"err", err)
		return err
	}

	return nil
}
