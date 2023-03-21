package main

import (
	"context"
	"crypto/tls"
	"github.com/kokizzu/gotro/L"
	"github.com/rocketlaunchr/https-go"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func main() {
	go func() {
		time.Sleep(2 * time.Second)
		ctx := context.Background()

		opts := []func(options *config.LoadOptions) error{}
		opts = append(opts, config.WithRegion("us-east-1"))

		cred := credentials.NewStaticCredentialsProvider("notImportanKey1", "notImportantSecret1", "")
		opts = append(opts, config.WithCredentialsProvider(cred))
		var awsCfg aws.Config

		resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...any) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:       "aws",
				SigningRegion:     region,
				URL:               `https://localhost:8080`,
				HostnameImmutable: true,
			}, nil
		})
		opts = append(opts, config.WithEndpointResolverWithOptions(resolver))
		opts = append(opts, config.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}))

		awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
		L.PanicIf(err, `config.LoadDefaultConfig`)

		s3Client := s3.NewFromConfig(awsCfg)

		bucketName := `bucket1`
		key := `key1`
		const content = "Hello world\n123\n"

		_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:            &bucketName,
			Key:               &key,
			Body:              strings.NewReader(content),
			ChecksumAlgorithm: s3types.ChecksumAlgorithmCrc32,
		})
		L.PanicIf(err, `s3Client.PutObject`)
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		log.Printf("[%s]\n", string(b))
		w.WriteHeader(http.StatusNoContent)
	})

	httpServer, _ := https.Server("8080", https.GenerateOptions{Host: "localhost"})
	log.Fatal(httpServer.ListenAndServeTLS("", ""))
}
