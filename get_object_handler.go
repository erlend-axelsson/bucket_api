package main

import (
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func getObjectHandler(w http.ResponseWriter, r *http.Request) {
	defer closeRequest(r)
	qVals := r.URL.Query()
	key := strings.TrimSpace(r.PathValue("key"))
	if key == "" {
		errorHandler(w, r, http.StatusBadRequest, "key is required, ex /get/directory/subdirectory/object.txt")
		return
	}
	var etag *string
	_etag := strings.TrimSpace(qVals.Get("etag"))
	if _etag != "" {
		etag = &_etag
	}

	var modSince *time.Time
	_modSince := qVals.Get("modified_since")
	if _modSince != "" {
		t, err := time.Parse(_modSince, http.TimeFormat)
		if err == nil {
			modSince = &t
		}
	}
	obj, err := getObject(key, etag, modSince)
	if err != nil {
		slog.Error(err.Error())
		errorHandler(w, r, http.StatusInternalServerError, "api error, check the logs if you have access")
		return
	}
	w.Header().Set("Bucket-Key", obj.Key)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName(obj.Key)+"\"")
	if obj.ETag != nil {
		w.Header().Set("ETag", *obj.ETag)
	}
	if obj.LastModified != nil {
		w.Header().Set("Last-Modified", httpTimeString(*obj.LastModified))
	}
	if obj.ContentType != nil {
		w.Header().Set("Content-Type", *obj.ContentType)
	}
	if _, err = w.Write(obj.Content); err != nil {
		slog.Error(err.Error())
		fallbackErrorHandler(w, r)
	}
}
