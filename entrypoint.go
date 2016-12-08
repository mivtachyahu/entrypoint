package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func runExec(executable string, args []string) {
	err := exec.Command(executable, args...).Run()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Executed command", executable, args)
}

func getMyRegion() string {
	region, err := ec2metadata.New(session.New()).Region()
	if err != nil {
		log.Printf("Unable to retrieve the region from the EC2 instance %v\n", err)
	}
	return region
}

func getBucketRegion(bucketName string) *string {
	sess, err := session.NewSession()
	if err != nil {
		log.Fatal("failed to create session,", err)
	}

	svc := s3.New(sess)

	params := &s3.GetBucketLocationInput{
		Bucket: aws.String(bucketName),
	}
	resp, err := svc.GetBucketLocation(params)

	if err != nil {
		log.Fatal(err.Error())
	}
	return resp.LocationConstraint
}

func getFile(fileName, bucketName, keyName string) {
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal("Failed to create file", err)
	}
	defer file.Close()
	bucketRegion := getBucketRegion(bucketName)
	downloader := s3manager.NewDownloader(session.New(&aws.Config{Region: bucketRegion}))
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(keyName),
		})
	if err != nil {
		log.Println("Failed to download file", err)
		return
	}

	log.Println("Downloaded file", file.Name(), numBytes, "bytes")
}

func main() {
	bucketPtr := flag.String("bucket", "", "bucket where the file is located")
	keyPtr := flag.String("key", "", "key within the bucket")
	destPtr := flag.String("dest", "", "destination to store the file")
	execPtr := flag.String("exec", "", "executable and any arguments to run")
	log.Println("Entrypoint for Docker Containers on AWS")
	flag.Parse()
	if *bucketPtr == "" || *keyPtr == "" || *destPtr == "" {
		flag.PrintDefaults()
		log.Fatal("Required Flag Not Given")
	}
	command := strings.Split(*execPtr, " ")[0]
	args := strings.Split(*execPtr, " ")[1:]
	getFile(*destPtr, *bucketPtr, *keyPtr)
	runExec(command, args)
}
