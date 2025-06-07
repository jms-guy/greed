package handlers

import (
	"encoding/json"
	"net/http"
)

func (app *AppServer) respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		app.Logger.Log(
			"level", "error",
			"status code", code,
			"msg", msg,
			"err", err,
		)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	app.respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func (app *AppServer) respondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		app.Logger.Log(
			"level", "error",
			"msg", "error marshalling JSON",
			"err", err,
		)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}