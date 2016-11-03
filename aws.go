package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"os"
	"path/filepath"
)

var AWSSession *session.Session

func initAWS() {
	defer Track("initAWS", Now(), debugOut)

	AWSSession = session.New()

	// Region
	if GlobalConfig.Get("awsRegion") != "" {
		// CLI trumps
		AWSSession.Config.Region = aws.String(GlobalConfig.Get("awsRegion"))
	} else if os.Getenv("AWS_REGION") == "" {
		// Grab it from this EC2 instace
		region, err := ec2metadata.New(session.New()).Region()
		if err != nil {
			fmt.Printf("Cannot set AWS region: '%v'\n", err)
			os.Exit(1)
		}
		AWSSession.Config.Region = aws.String(region)
	}

	// Creds
	if GlobalConfig.Get("awsAccessKey") != "" && GlobalConfig.Get("awsSecretKey") != "" {
		// CLI trumps
		creds := credentials.NewStaticCredentials(
			GlobalConfig.Get("awsAccessKey"),
			GlobalConfig.Get("awsSecretKey"),
			"")
		AWSSession.Config.Credentials = creds
	}

}

func fileToBucket(filename, bucket string) (size int64, err error) {
	defer Track("fileToBucket", Now(), debugOut)

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	// Get the filesize
	fi, ferr := file.Stat()
	if ferr == nil {
		size = fi.Size()
	}

	// Extract the basename
	baseFilename := filepath.Base(filename)

	// Setup the uploader, and git'r'done
	svc := s3manager.NewUploader(AWSSession)
	_, err = svc.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(baseFilename),
		Body:   file,
	})

	return
}
