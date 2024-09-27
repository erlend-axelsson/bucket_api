package main

import (
	"bytes"
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type contentInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	Etag         string
}

type Object struct {
	Key          string
	ETag         *string
	LastModified *time.Time
	ContentType  *string
	Content      []byte
}

type Article struct {
	ArticleKey    string   `json:"articleKey,omitempty"`
	ArticleAssets []string `json:"articleAssets,omitempty"`
}

func listArticles() ([]Article, error) {
	content, err := listObjects("/Articles")
	if err != nil {
		return nil, err
	}
	var buckets = make(map[string]Article)
	for _, info := range content {
		dirName, filename := filepath.Split(info.Key)
		if filename != "" {
			article, ok := buckets[dirName]
			if !ok {
				article = Article{}
			}
			if strings.HasSuffix(filename, ".html") {
				article.ArticleKey = info.Key
			} else {
				article.ArticleAssets = append(article.ArticleAssets, article.ArticleKey)
			}
			buckets[dirName] = article
		}
	}
	var out []Article
	for _, article := range buckets {
		out = append(out, article)
	}
	return out, err
}

func listObjects(prefix string) ([]contentInfo, error) {
	var pf *string = nil
	if prefix != "" {
		pf = &prefix
	}
	return doListObjects(pf, nil, make([]contentInfo, 0, 20))
}
func doListObjects(prefix *string, continuationToken *string, accumulator []contentInfo) ([]contentInfo, error) {
	inputArgs := &s3.ListObjectsV2Input{
		Bucket: aws.String(envOpts.bucketName),
	}
	if prefix != nil {
		inputArgs.Prefix = prefix
	}
	if continuationToken != nil {
		inputArgs.ContinuationToken = continuationToken
	}

	output, err := client.ListObjectsV2(context.TODO(), inputArgs)
	if err != nil {
		return nil, err
	}
	if output == nil {
		return nil, err
	}
	var out = accumulator
	for _, content := range output.Contents {
		out = append(out, contentInfo{
			Key:          aws.ToString(content.Key),
			Size:         aws.ToInt64(content.Size),
			LastModified: aws.ToTime(content.LastModified),
			Etag:         aws.ToString(content.ETag),
		})
	}
	if output.IsTruncated != nil && *output.IsTruncated {
		return doListObjects(prefix, output.NextContinuationToken, out)
	}
	return out, nil
}

func getObject(key string, etag *string, modifiedSince *time.Time) (Object, error) {
	buffer := bytes.Buffer{}
	out := Object{Key: key}
	obj, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Key:             &key,
		Bucket:          &envOpts.bucketName,
		IfNoneMatch:     etag,
		IfModifiedSince: modifiedSince,
	})
	if err != nil {
		return out, err
	}
	if obj == nil {
		return out, errors.New("s3 object not found")
	}
	defer func(Body io.ReadCloser) {
		closeErr := Body.Close()
		if closeErr != nil {

		}
	}(obj.Body)

	_, err = buffer.ReadFrom(obj.Body)
	if err != nil {
		return out, err
	}

	out.ETag = obj.ETag
	out.LastModified = modifiedSince
	out.ContentType = obj.ContentType
	out.Content = buffer.Bytes()
	return out, nil
}

func putObject(dir, fileName string, b []byte, mime *string) (*string, error) {
	fileKey := path.Join(dir, fileName)
	res, err := client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      &envOpts.bucketName,
		Key:         &fileKey,
		Body:        bytes.NewBuffer(b),
		ContentType: mime,
	})
	if res == nil {
		return nil, err
	}
	return res.ETag, err
}

func deleteObject(dir, fileName string) error {
	fileKey := path.Join(dir, fileName)
	_, err := client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Key:    &fileKey,
		Bucket: &envOpts.bucketName,
	})
	return err
}

func fileName(filePath string) string {
	_, filename := path.Split(filePath)
	return filename
}
