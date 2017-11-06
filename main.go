package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"./src/bucket"
	"./src/logger"
)

func runExec(executable string, args []string) {
	logger.Trace.Println("runExec Function")
	err := exec.Command(executable, args...).Run()
	if err != nil {
		logger.Error.Fatal(err)
	}
	logger.Info.Println("Executed command", executable, args)
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
		logger.LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	} else {
		logger.LogInit(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	}

	logger.Info.Println("Entrypoint for Docker Containers on AWS")
	if *bucketPtr == "" || *keyPtr == "" || *destPtr == "" {
		flag.PrintDefaults()
		logger.Error.Fatal("Required Flag Not Given")
	}
	bucket.GetFile(*destPtr, *bucketPtr, *cachePtr, *keyPtr)
	command := strings.Split(*execPtr, " ")[0]
	args := strings.Split(*execPtr, " ")[1:]
	if command != "" {
		runExec(command, args)
	}
}
