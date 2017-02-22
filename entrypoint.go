package main

import (
	"flag"
	"fmt"
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

func getFileSize(fileName string) int64 {
	file, err := os.Open(fileName)
	if err != nil {
		Error.Fatal("Failed to open file", err)
	}
	defer file.Close()

	stats, statsErr := file.Stat()
	if statsErr != nil {
		Error.Fatal("Failed to stat file", statsErr)
	}
	size := stats.Size()
	return size
}

func getObjectDetails(bucketName, keyName string) *s3.HeadObjectOutput {
	Trace.Println("getObjectDetails Function")
	bucketRegion := getBucketRegion(bucketName)
	sess, err := session.NewSession(&aws.Config{Region: bucketRegion})
	if err != nil {
		Error.Fatal("failed to create session,", err)
	}

	svc := s3.New(sess)
	params := &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(keyName),
	}
	resp, err := svc.HeadObject(params)

	if err != nil {
		Error.Fatal(err.Error())
	}
	return resp
}

func downloadFile(fileName, bucketName, keyName string) {
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

func copyFile(src, dst string) (int64, error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	srcFileStat, err := srcFile.Stat()
	if err != nil {
		return 0, err
	}

	if !srcFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer dstFile.Close()
	return io.Copy(dstFile, srcFile)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func getFile(fileName, bucketName, cacheName, keyName string) {
	Trace.Println("getFile Function")
	localFile := strings.Join([]string{cacheName, keyName}, "")
	localPath := strings.Join(strings.Split(localFile, "/")[:len(strings.Split(localFile, "/"))-1], "/")
	if !exists(localPath) {
		Trace.Println("Cache folder doesn't exist - creating")
		createDir(localPath)
	}
	if !exists(fileName) {
		Trace.Println("File doesn't exist - creating")
		downloadFile(localFile, bucketName, keyName)
		Trace.Println("File Downloaded - Copying to destination")
		copyFile(localFile, fileName)
	} else {
		Trace.Println("File exists - comparing sizes")
		if *getObjectDetails(bucketName, keyName).ContentLength == getFileSize(fileName) {
			Trace.Println("Sizes Match - not downloading")
		} else {
			Trace.Println("Sizes Do Not Match - Downloading")
			downloadFile(localFile, bucketName, keyName)
			Trace.Println("File Downloaded - Copying to destination")
			copyFile(localFile, fileName)
		}
	}
}

func createDir(path string) {
	err := os.MkdirAll(path, 0644)
	if err != nil {
		Error.Fatal("Failed to create  dir", err)
	}
	Info.Println("Created folder", path)
}

func main() {
	bucketPtr := flag.String("bucket", "", "bucket where the file is located")
	keyPtr := flag.String("key", "", "key within the bucket")
	destPtr := flag.String("dest", "", "destination to store the file")
	cachePtr := flag.String("cache", "/var/cache/download", "destination to store the file temporarily")
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
	getFile(*destPtr, *bucketPtr, *cachePtr, *keyPtr)
	if command != "" {
		runExec(command, args)
	}
}
