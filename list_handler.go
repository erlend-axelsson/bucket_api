package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ResponseObject struct {
	Key          string `json:"bucket_key"`
	FileName     string `json:"file_name"`
	Size         int64  `json:"size"`
	Etag         string `json:"etag"`
	LastModified string `json:"last_modified"`
}

type ListResponse struct {
	HttpStatus int              `json:"status"`
	NumObjects int              `json:"num_objects"`
	Content    []ResponseObject `json:"content"`
}

func listObjectsHandler(w http.ResponseWriter, r *http.Request) {
	defer closeRequest(r)
	vals := r.URL.Query()
	content, err := listObjects(vals.Get("prefix"))
	if err != nil {
		fallbackErrorHandler(w, r)
		return
	}
	outContent := make([]ResponseObject, 0, len(content))
	for _, info := range content {
		outContent = append(
			outContent,
			ResponseObject{
				Key:          info.Key,
				FileName:     fileName(info.Key),
				Size:         info.Size,
				Etag:         info.Etag,
				LastModified: httpTimeString(info.LastModified),
			})
	}
	resp := &ListResponse{
		HttpStatus: http.StatusOK,
		NumObjects: len(outContent),
		Content:    outContent,
	}
	outBytes, err := json.MarshalIndent(resp, "", "\t")
	if err != nil {
		slog.Error(err.Error())
		fallbackErrorHandler(w, r)
	} else {
		w.Header().Set(ContentTypeKey, ApplicationJSON)
		w.WriteHeader(http.StatusOK)
		if _, err = w.Write(outBytes); err != nil {
			slog.Error(err.Error())
			fallbackErrorHandler(w, r)
		}
	}
}
