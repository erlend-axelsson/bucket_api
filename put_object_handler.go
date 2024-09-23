package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const FileLimitBytes = 2 << 19

const ContentDispositionHeader = "Content-Disposition"
const ContentTypeHeader = "Content-Type"
const ContentLengthHeader = "Content-Length"
const BucketPrefixHeader = "Bucket-Prefix"

var dispositionExpression = regexp.MustCompile(`filename="(?P<Filename>.*)"`)
var filenameIndex = dispositionExpression.SubexpIndex("Filename")

func ParseDisposition(dispositionHeader string) string {
	matches := dispositionExpression.FindStringSubmatch(dispositionHeader)
	if len(matches) <= filenameIndex {
		return ""
	}
	return matches[filenameIndex]
}

func validateHeaders(disposition, size string) (string, error) {
	filename, err := validateDisposition(strings.TrimSpace(disposition))
	if err != nil {
		return "", err
	}
	_, err = validateSize(strings.TrimSpace(size))
	if err != nil {
		return "", err
	}
	return filename, nil
}

func validateDisposition(disposition string) (string, error) {
	filename := ParseDisposition(disposition)
	if filename == "" {
		return "", errors.New("invalid disposition, filename is empty")
	}
	return filename, nil
}

func validateSize(size string) (int, error) {
	parsedSize, err := strconv.Atoi(size)
	if err != nil {
		return -1, err
	}
	if parsedSize < 1 || parsedSize > FileLimitBytes {
		return -1, fmt.Errorf("size must be between 1 and %d", FileLimitBytes)
	}
	return parsedSize, nil
}

func putObjectHandler(w http.ResponseWriter, r *http.Request) {
	defer closeRequest(r)
	filename, err := validateHeaders(
		r.Header.Get(ContentDispositionHeader),
		r.Header.Get(ContentLengthHeader))
	if err != nil {
		errorHandler(w, r, http.StatusBadRequest, err.Error())
		return
	}
	bucketPrefix := r.Header.Get(BucketPrefixHeader)
	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		errorHandler(w, r, http.StatusBadRequest, "unable to read request body")
		return
	}

	fileBytes := buf.Bytes()
	mime := strings.TrimSpace(r.Header.Get(ContentTypeHeader))
	if mime == "" {
		mime = http.DetectContentType(fileBytes)
	}

	etag, err := putObject(bucketPrefix, filename, fileBytes, &mime)
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, "Something went wrong when uploading object")
		return
	}

	if etag != nil {
		w.Header().Set("ETag", *etag)
	}
	w.WriteHeader(http.StatusNoContent)
}
