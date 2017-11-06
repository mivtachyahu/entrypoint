package bucket

import (
	"io/ioutil"
	"os"
	"strings"

	"../fs"
	"../logger"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func getObjectDetails(bucketName, keyName string) *s3.HeadObjectOutput {
	logger.Trace.Println("getObjectDetails Function")
	bucketRegion := getBucketRegion(bucketName)
	sess, err := session.NewSession(&aws.Config{Region: bucketRegion})
	if err != nil {
		logger.Error.Fatal("failed to create session,", err)
	}

	svc := s3.New(sess)
	params := &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(keyName),
	}
	resp, err := svc.HeadObject(params)

	if err != nil {
		logger.Error.Fatal(err.Error())
	}
	return resp
}

func getMyRegion() string {
	logger.Trace.Println("getMyRegion Function")
	region, err := ec2metadata.New(session.New()).Region()
	if err != nil {
		logger.Warning.Printf("Unable to retrieve the region from the EC2 instance %v\n", err)
	}
	return region
}

func getBucketRegion(bucketName string) *string {
	logger.Trace.Println("getBucketRegion Function")
	sess, err := session.NewSession(&aws.Config{Region: aws.String("eu-west-1")})
	if err != nil {
		logger.Error.Fatal("failed to create session,", err)
	}

	svc := s3.New(sess)

	params := &s3.GetBucketLocationInput{
		Bucket: aws.String(bucketName),
	}
	resp, err := svc.GetBucketLocation(params)

	if err != nil {
		logger.Error.Fatal(err.Error())
	}
	return resp.LocationConstraint
}

func downloadFile(fileName, bucketName, keyName string) {
	logger.Trace.Println("downloadFile Function")
	file, err := ioutil.TempFile(os.TempDir(), "prefix")
	defer os.Remove(file.Name())
	logger.Trace.Println("Created file")
	bucketRegion := getBucketRegion(bucketName)
	downloader := s3manager.NewDownloader(session.New(&aws.Config{Region: bucketRegion}))
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(keyName),
		})
	if err != nil {
		logger.Error.Println("Failed to download file", err)
		return
	}
	fs.CopyFile(file.Name(), fileName)
	logger.Info.Println("Downloaded file", fileName, numBytes, "bytes")
}

// GetFile fileName (destination) bucketName (source Bucket) cacheName (local place to cache files) keyName (source Key in Bucket)
func GetFile(fileName, bucketName, cacheName, keyName string) {
	logger.Trace.Println("getFile Function")
	localFile := strings.Join([]string{cacheName, keyName}, "")
	localPath := strings.Join(strings.Split(localFile, "/")[:len(strings.Split(localFile, "/"))-1], "/")
	if !fs.Exists(localPath) {
		logger.Trace.Println("Cache folder doesn't exist - creating")
		fs.CreateDir(localPath)
	}
	if !fs.Exists(localFile) {
		logger.Trace.Println("Cache File doesn't exist - creating")
		downloadFile(localFile, bucketName, keyName)
		logger.Trace.Println("Cache File Downloaded - Copying to destination")
		fs.CopyFile(localFile, fileName)
	} else {
		logger.Trace.Println("File exists - comparing sizes")
		if *getObjectDetails(bucketName, keyName).ContentLength == fs.GetFileSize(localFile) {
			logger.Trace.Println("Sizes Match - Copying Cached to destination")
			fs.CopyFile(localFile, fileName)
		} else {
			logger.Trace.Println("Sizes Do Not Match - Downloading")
			downloadFile(localFile, bucketName, keyName)
			logger.Trace.Println("File Downloaded - Copying to destination")
			fs.CopyFile(localFile, fileName)
		}
	}
}
