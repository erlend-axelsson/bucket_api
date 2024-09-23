package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

const ContentTypeKey = "Content-Type"
const ApplicationJSON = "application/json; charset=utf-8"

const internalServerError = "{\n  \"status\": 500,\n  \"message\": \"Something went wrong\"\n}"

type jsonErrResp struct {
	HttpStatus int    `json:"status"`
	Message    string `json:"message"`
}

func errorJson(httpStatus int, message string) ([]byte, error) {
	jsonErr := jsonErrResp{
		HttpStatus: httpStatus,
		Message:    message,
	}
	data, err := json.Marshal(&jsonErr)
	if err != nil {
		return []byte(internalServerError), err
	}
	return data, nil
}

func closeRequest(r *http.Request) {
	if err := r.Body.Close(); err != nil {
		slog.Error(err.Error())
	}
}

func errorHandler(w http.ResponseWriter, r *http.Request, status int, message string) {
	errorBytes, err := errorJson(status, message)
	if err != nil {
		fallbackErrorHandler(w, r)
		return
	}

	w.Header().Set(ContentTypeKey, ApplicationJSON)
	w.WriteHeader(status)
	_, err = w.Write(errorBytes)
	if err != nil {
		fallbackErrorHandler(w, r)
	}
}

func fallbackErrorHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(ContentTypeKey, ApplicationJSON)
	w.WriteHeader(http.StatusInternalServerError)
	_, err := w.Write([]byte(internalServerError))
	if err != nil {
		slog.Error("Error writing error response: %s", err)
	}
}

func httpTimeString(t time.Time) string {
	return t.UTC().Format(http.TimeFormat)
}
