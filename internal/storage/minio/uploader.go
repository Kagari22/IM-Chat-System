package minio

import (
	"context"
	"fmt"
	"io"
	"strings"

	miniogo "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"IM_Chat_System/internal/storage"
)

type Uploader struct {
	client        *miniogo.Client
	bucket        string
	publicBaseURL string
}

func New(endpoint, accessKey, secretKey, bucket string, useSSL bool, publicBaseURL string) (*Uploader, error) {
	client, err := miniogo.New(endpoint, &miniogo.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	u := &Uploader{
		client:        client,
		bucket:        bucket,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, miniogo.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}
	if err := client.SetBucketPolicy(ctx, bucket, buildPublicReadPolicy(bucket)); err != nil {
		return nil, err
	}
	return u, nil
}

func (u *Uploader) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (storage.ObjectInfo, error) {
	info, err := u.client.PutObject(ctx, u.bucket, objectName, reader, size, miniogo.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return storage.ObjectInfo{}, err
	}

	url := fmt.Sprintf("%s/%s/%s", u.publicBaseURL, u.bucket, objectName)
	return storage.ObjectInfo{
		Key:         objectName,
		URL:         url,
		ContentType: contentType,
		Size:        info.Size,
	}, nil
}

func buildPublicReadPolicy(bucket string) string {
	return fmt.Sprintf(`{
  "Version":"2012-10-17",
  "Statement":[
    {
      "Effect":"Allow",
      "Principal":{"AWS":["*"]},
      "Action":["s3:GetBucketLocation","s3:ListBucket"],
      "Resource":["arn:aws:s3:::%s"]
    },
    {
      "Effect":"Allow",
      "Principal":{"AWS":["*"]},
      "Action":["s3:GetObject"],
      "Resource":["arn:aws:s3:::%s/*"]
    }
  ]
}`, bucket, bucket)
}
