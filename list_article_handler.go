package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ArticlesResponse struct {
	HttpStatus  int       `json:"httpStatus"`
	Content     []Article `json:"content"`
	NumArticles int       `json:"numArticles"`
}

func listArticlesHandler(w http.ResponseWriter, r *http.Request) {
	defer closeRequest(r)
	content, err := listArticles()
	if err != nil {
		fallbackErrorHandler(w, r)
		return
	}
	resp := &ArticlesResponse{
		HttpStatus:  http.StatusOK,
		NumArticles: len(content),
		Content:     content,
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
