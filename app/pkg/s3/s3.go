package s3

import (
	"context"
	"fmt"
	"github.com/ahmadateya/flotta-webapp-backend/config"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io/ioutil"
	"log"
)

// This package is used to encapsulate the S3 treatment

type S3 struct {
	awsClient *s3.Client
	bucket    string
}

// InitS3Client initializes the S3 client
func InitS3Client() S3 {
	// read configurations from env file
	cfg, _ := config.NewConfig("./config.yaml")

	awsCfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion(cfg.Storage.S3.Region),
		awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.Storage.S3.AccessKeyId,
				cfg.Storage.S3.SecretAccessKey,
				"")),
	)
	if err != nil {
		log.Fatal(err)
	}

	client := s3.NewFromConfig(awsCfg)
	return S3{
		awsClient: client,
		bucket:    cfg.Storage.S3.Bucket,
	}
}

// ListTopLevelFolders returns the list of top level folders in the S3 bucket
// Top Level folders is supposed to represent the machines registered in Flotta
func (s *S3) ListTopLevelFolders() []string {
	delimiter := "/"
	resp, err := s.awsClient.ListObjectsV2(
		context.TODO(),
		&s3.ListObjectsV2Input{
			Bucket:    &s.bucket,
			Delimiter: &delimiter,
		})

	if err != nil {
		fmt.Printf("Got error retrieving list of objects: %v\n", err)
	}

	topLevelFolders := make([]string, 0)
	for _, obj := range resp.CommonPrefixes {
		topLevelFolders = append(topLevelFolders, *obj.Prefix)
	}

	return topLevelFolders
}

func (s *S3) GetMostRecentObjectNameInFolder(folder string) string {
	folderPath := folder + "/"
	resp, err := s.awsClient.ListObjectsV2(
		context.TODO(),
		&s3.ListObjectsV2Input{
			Bucket: &s.bucket,
			Prefix: &folderPath,
		})

	if err != nil {
		fmt.Printf("Got error retrieving list of objects: %v\n", err)
	}

	if len(resp.Contents) > 0 {
		return *resp.Contents[1].Key
	}

	return ""
}

func (s *S3) ReadObject(objectPath string) string {
	input := &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &objectPath,
	}

	resp, err := s.awsClient.GetObject(context.TODO(), input)
	if err != nil {
		fmt.Printf("Got error retrieving object: %v\n", err)
	}
	defer resp.Body.Close()

	objContent, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	return fmt.Sprintf("%s", objContent)
}