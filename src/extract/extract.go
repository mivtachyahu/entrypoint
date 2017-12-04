package extract

func unGzip(source, target string) {
	Trace.Println("unGzip Function")
	reader, err := os.Open(source)
	if err != nil {
		Error.Fatal("Failed to open source", err)
	}
	defer reader.Close()

	archive, err := gzip.NewReader(reader)
	if err != nil {
		Error.Fatal("Failed to open archive contents", err)
	}
	defer archive.Close()

	target = filepath.Join(target, archive.Name)
	writer, err := os.Create(target)
	if err != nil {
		Error.Fatal("Failed to open destination", err)
	}
	defer writer.Close()

	_, err = io.Copy(writer, archive)
	if err != nil {
		Error.Fatal("Failed to copy contents to destination", err)
	}
}

func unTar(tarball, target string) {
	Trace.Println("unTar Function")
	reader, err := os.Open(tarball)
	if err != nil {
		Error.Fatal("Failed to read tarball", err)
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			Trace.Println("End of Tarball Reached - Finished Extracting")
			break
		} else if err != nil {
			Error.Fatal("Failed read next block of tarball", err)
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				Error.Fatal("Failed to make directory", err)
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			Error.Fatal("Failed to create file to write", err)
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			Error.Fatal("Failed to copy from tarball to file", err)
		}
	}
}
func dirName(wholeFileName string) string {

}


func unextract(wholeFileName string) {
	Trace.Println("unextract Function")
	fileName := strings.Split(wholeFileName, "/")[len(strings.Split(wholeFileName))-1]
	filePath := strings.Join()
	fileSplit := strings.Split(fileName, ".")
	switch fileEnd := fileSplit[len(fileSplit)-1]; fileEnd {
	case "tar":
		Trace.Println("detected tgz")
		unTar(wholeFileName, destination)
	case "tgz":
		Trace.Println("detected tgz")
		unGzip(wholeFileName, tempDestination)
		// break down wholeFileName with .tar on the end
		unTar()
	case "gz":
		// lets check if it's a tar.gz
		switch oneBack := fileSplit[len(fileSplit)-2]; oneBack {
		case "tar":
			Trace.Println("detected tar gz")
			unGzip()
			unTar()
		default:
			Trace.Println("detected gzip")
			unGzip()
		}
	default:
		Info.Println("Not detected tar / tgz / tar.gz")
	}
}
