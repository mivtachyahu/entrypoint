package main

import (
	"flag"
	"io"
	"io/ioutil"
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

var (
	// Trace - Trace Logging
	Trace *log.Logger
	// Info - Info Logging
	Info *log.Logger
	// Warning - Warning Logging
	Warning *log.Logger
	// Error - Error Logging
	Error *log.Logger
)

// LogInit - Initialise the logging styles
func LogInit(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {
	Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

func runExec(executable string, args []string) {
	Trace.Println("runExec Function")
	err := exec.Command(executable, args...).Run()
	if err != nil {
		Error.Fatal(err)
	}
	Info.Println("Executed command", executable, args)
}

func getMyRegion() string {
	Trace.Println("getMyRegion Function")
	region, err := ec2metadata.New(session.New()).Region()
	if err != nil {
		Warning.Printf("Unable to retrieve the region from the EC2 instance %v\n", err)
	}
	return region
}

func getBucketRegion(bucketName string) *string {
	Trace.Println("getBucketRegion Function")
	sess, err := session.NewSession(&aws.Config{Region: aws.String("eu-west-1")})
	if err != nil {
		Error.Fatal("failed to create session,", err)
	}

	svc := s3.New(sess)

	params := &s3.GetBucketLocationInput{
		Bucket: aws.String(bucketName),
	}
	resp, err := svc.GetBucketLocation(params)

	if err != nil {
		Error.Fatal(err.Error())
	}
	return resp.LocationConstraint
}

func getFile(fileName, bucketName, keyName string) {
	Trace.Println("getFile Function")
	file, err := os.Create(fileName)
	if err != nil {
		Error.Fatal("Failed to create file", err)
	}
	defer file.Close()
	Trace.Println("Created file")
	bucketRegion := getBucketRegion(bucketName)
	downloader := s3manager.NewDownloader(session.New(&aws.Config{Region: bucketRegion}))
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(keyName),
		})
	if err != nil {
		Error.Println("Failed to download file", err)
		return
	}

	Info.Println("Downloaded file", file.Name(), numBytes, "bytes")
}

func main() {
	bucketPtr := flag.String("bucket", "", "bucket where the file is located")
	keyPtr := flag.String("key", "", "key within the bucket")
	destPtr := flag.String("dest", "", "destination to store the file")
	execPtr := flag.String("exec", "", "executable and any arguments to run")
	verbosePtr := flag.Bool("v", false, "Verbose logging")
	flag.Parse()
	if *verbosePtr {
		LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	} else {
		LogInit(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	}
	Info.Println("Entrypoint for Docker Containers on AWS")
	if *bucketPtr == "" || *keyPtr == "" || *destPtr == "" {
		flag.PrintDefaults()
		Error.Fatal("Required Flag Not Given")
	}
	command := strings.Split(*execPtr, " ")[0]
	args := strings.Split(*execPtr, " ")[1:]
	getFile(*destPtr, *bucketPtr, *keyPtr)
	if command != "" {
		runExec(command, args)
	}
}
