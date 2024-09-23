package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

var routeMap = map[string]string{
	"/list":            "GET list objects",
	"/get/{key...}":    "GET get object",
	"/put":             "PUT put object",
	"/delete/{key...}": "DELETE delete object",
}
var routeBytes = marshalRouteInfo(routeMap)

func marshalRouteInfo(routeInfo map[string]string) []byte {
	rBytes, err := json.MarshalIndent(&routeInfo, "", "\t")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	return rBytes
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(ContentTypeKey, ApplicationJSON)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(routeBytes); err != nil {
		fallbackErrorHandler(w, r)
	}
}
