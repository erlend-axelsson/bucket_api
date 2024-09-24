package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type statusResponse struct {
	IsAlive   bool   `json:"isAlive"`
	IsHealthy bool   `json:"isHealthy"`
	Started   string `json:"started"`
}

var status []byte

func isAliveHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(status); err != nil {
		errorHandler(
			w,
			r,
			http.StatusInternalServerError,
			"failed to output status response, not healthy I guess",
		)
	}
}

func init() {
	sr := statusResponse{
		IsAlive:   true,
		IsHealthy: true,
		Started:   time.Now().Format(http.TimeFormat),
	}
	statusBytes, err := json.MarshalIndent(&sr, "", "\t")
	if err != nil {
		panic(err)
	}
	status = statusBytes
}
