package main

import (
	"net/http"
	"path"
	"strings"
)

func deleteObjectHandler(w http.ResponseWriter, r *http.Request) {
	closeRequest(r)
	key := strings.TrimSpace(r.PathValue("key"))
	if key == "" {
		errorHandler(w, r, http.StatusNotFound, "key is required")
		return
	}
	dir, file := path.Split(key)
	err := deleteObject(dir, file)
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, "could not delete")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
