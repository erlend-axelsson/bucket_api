package main

import (
	"context"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"
	"log/slog"
	"net/http"
	"os"
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

var client *s3.Client
var envOpts = initOptions()
var allowInsecure = false

func getCertsFromEnv() (*tls.Config, error) {
	base64Ca := os.Getenv("CA_CERT")
	base64ServerCrt := os.Getenv("SERVER_CERT")
	base64ServerKey := os.Getenv("SERVER_KEY")

	rawCa, err := base64.StdEncoding.DecodeString(base64Ca)
	if err != nil {
		return nil, err
	}
	rawServerCrt, err := base64.StdEncoding.DecodeString(base64ServerCrt)
	if err != nil {
		return nil, err
	}
	rawServerKey, err := base64.StdEncoding.DecodeString(base64ServerKey)
	if err != nil {
		return nil, err
	}

	caCrt, err := x509.ParseCertificate(rawCa)
	if err != nil {
		return nil, err
	}
	serverCrt, err := x509.ParseCertificate(rawServerCrt)
	if err != nil {
		return nil, err
	}
	privateKey := parsePk(rawServerKey, serverCrt.PublicKeyAlgorithm)
	if privateKey == nil {
		return nil, fmt.Errorf("unable to parse private key")
	}
	crtPool := x509.NewCertPool()
	crtPool.AddCert(caCrt)
	tlsCrt := tls.Certificate{
		Certificate: [][]byte{rawServerCrt, rawCa},
		PrivateKey:  privateKey,
		Leaf:        serverCrt,
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCrt},
		RootCAs:      crtPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    crtPool,
	}, nil

}

func parsePk(raw []byte, alg x509.PublicKeyAlgorithm) crypto.PrivateKey {
	switch alg {
	case x509.RSA:
		pk, err := x509.ParsePKCS1PrivateKey(raw)
		if err != nil {
			return nil
		}
		return pk
	case x509.ECDSA:
		pk, err := x509.ParseECPrivateKey(raw)
		if err != nil {
			return nil
		}
		return pk
	case x509.Ed25519:
		pk, err := x509.ParsePKCS1PrivateKey(raw)
		if err != nil {
			return nil
		}
		return pk
	default:
		return nil
	}

}

func checkAllowInsecure() bool {
	ans := os.Getenv("ALLOW_INSECURE")
	if ans == "true" {
		allowInsecure = true
	}
	return allowInsecure
}

func initServ() *http.Server {
	var tlsConfig *tls.Config = nil
	if !checkAllowInsecure() {
		conf, err := getCertsFromEnv()
		if err != nil {
			slog.Error("Error getting cert from env")
			os.Exit(1)
		}
		tlsConfig = conf
	}

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
	mux.HandleFunc("GET /list", listObjectsHandler)
	mux.HandleFunc("GET /get/{key...}", getObjectHandler)
	mux.HandleFunc("PUT /put", putObjectHandler)
	mux.HandleFunc("DELETE /delete/{key...}", deleteObjectHandler)
	return &http.Server{
		Handler:   mux,
		Addr:      ":5000",
		TLSConfig: tlsConfig,
	}
}

func main() {
	srv := initServ()
	err := srv.ListenAndServe()
	if err != nil {
		slog.Error(err.Error())
	}
}
