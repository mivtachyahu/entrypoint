package fs

import (
	"fmt"
	"io"
	"os"

	"../logger"
)

func GetFileSize(fileName string) int64 {
	logger.Trace.Println("getFileSize Function")
	file, err := os.Open(fileName)
	if err != nil {
		logger.Error.Fatal("Failed to open file", err)
	}
	defer file.Close()

	stats, statsErr := file.Stat()
	if statsErr != nil {
		logger.Error.Fatal("Failed to stat file", statsErr)
	}
	size := stats.Size()
	return size
}

func CopyFile(src, dst string) (int64, Error) {
	logger.Trace.Println("copyFile Function")
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

func Exists(path string) bool {
	logger.Trace.Println("exists Function")
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func CreateDir(path string) {
	err := os.MkdirAll(path, 0644)
	if err != nil {
		logger.Error.Fatal("Failed to create  dir", err)
	}
	logger.Info.Println("Created folder", path)
}
