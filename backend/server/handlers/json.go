package handlers

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func (app *AppServer) respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		_ = app.Logger.Log(
			"level", "error",
			"status code", code,
			"msg", msg,
			"err", err,
		)
	}

	app.respondWithJSON(w, code, ErrorResponse{
		Error: msg,
	})
}

func (app *AppServer) respondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		_ = app.Logger.Log(
			"level", "error",
			"msg", "error marshalling JSON",
			"err", err,
		)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	_, err = w.Write(dat)
	if err != nil {
		_ = app.Logger.Log(
			"level", "error",
			"msg", "writing error",
			"err", err,
		)
		return
	}
}
