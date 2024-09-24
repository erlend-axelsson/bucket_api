package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

type s3opts struct {
	accessKeyID     string
	secretAccessKey string
	endpointUrl     string
	region          string
	bucketName      string
}

func initOptions() s3opts {
	return s3opts{
		accessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		secretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		endpointUrl:     os.Getenv("AWS_ENDPOINT_URL_S3"),
		region:          os.Getenv("AWS_REGION"),
		bucketName:      os.Getenv("BUCKET_NAME"),
	}
}

const minSecretLength = 128
const SharedSecretEnv = "SHARED_SECRET"
const AuthorizationHeader = "Authorization"

var client *s3.Client
var envOpts = initOptions()
var allowInsecure = false

func checkAllowInsecure() bool {
	ans := os.Getenv("ALLOW_INSECURE")
	if ans == "true" {
		allowInsecure = true
	}
	return allowInsecure
}

type verifySecretHandler struct {
	sharedSecret string
	mux          *http.ServeMux
}

func (mu *verifySecretHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(AuthorizationHeader) == mu.sharedSecret {
		mu.mux.ServeHTTP(w, r)
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
}

func initServ() *http.Server {
	var handler http.Handler

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(envOpts.endpointUrl)
		o.Region = "auto"
	})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", rootHandler)
	mux.HandleFunc("GET /isalive", isAliveHandler)
	mux.HandleFunc("GET /list", listObjectsHandler)
	mux.HandleFunc("GET /get/{key...}", getObjectHandler)
	mux.HandleFunc("PUT /put", putObjectHandler)
	mux.HandleFunc("DELETE /delete/{key...}", deleteObjectHandler)

	if checkAllowInsecure() {
		handler = mux
	} else {
		secret := os.Getenv(SharedSecretEnv)
		if len(secret) < minSecretLength {
			minLen := strconv.Itoa(minSecretLength)
			slog.Error("Invalid shared secret value, must be at least " + minLen + " bytes")
		}
		handler = &verifySecretHandler{
			sharedSecret: secret,
			mux:          mux,
		}
	}

	return &http.Server{
		Handler: handler,
		Addr:    ":8080",
	}
}

func main() {
	srv := initServ()
	err := srv.ListenAndServe()
	if err != nil {
		slog.Error(err.Error())
	}
}
