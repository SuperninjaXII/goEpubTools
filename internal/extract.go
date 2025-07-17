package internal

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type MetaData struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Date        string `xml:"date"`
	Author      string `xml:"creator"`
}
type Package struct {
	MetaData MetaData `xml:"metadata"`
}

func EditEpub(data *Package, epubFile string, destination string) {
	os.MkdirTemp("", "./temp")
	go Unzip(epubFile, "./temp")
	AddMetaData(data.MetaData.Title, data.MetaData.Description, data.MetaData.Date, data.MetaData.Author, "./temp")
	go Compress("./temp", destination)
	defer os.RemoveAll("./temp")
}

func ReadMetaData(filePath string) (*Package, error) {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		fmt.Println("error opening zip, here is err :", err)
		return nil, err
	}
	defer reader.Close()
	// processing through the zip contentt
	for _, file := range reader.File {
		if path.Base(file.Name) == "content.opf" {
			contents, err := file.Open()
			if err != nil {
				fmt.Println("here is err", err)
			}
			defer contents.Close()

			data, err := io.ReadAll(contents)
			if err != nil {
				fmt.Println(err)
			}
			var v Package
			err = xml.Unmarshal([]byte(data), &v)
			if err != nil {
				fmt.Println("err umarshaling", err)
			}
			return &v, nil
		}
	}
	return nil, nil
}

func Unzip(src string, destination string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		fmt.Println("this is error from reading", err)
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		destPath := filepath.Join(destination, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(destPath, os.ModePerm)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
			return err
		}

		srcFile, err := file.Open()
		if err != nil {
			return err
		}

		destFile, err := os.Create(destPath)
		if err != nil {
			srcFile.Close()
			return err
		}

		_, err = io.Copy(destFile, srcFile)
		srcFile.Close()
		destFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func findOPF(dir string) string {
	opfPath := ""
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && filepath.Base(path) == "content.opf" {
			opfPath = path
			return io.EOF // Stop walking
		}
		return nil
	})
	return opfPath
}

func AddMetaData(title, description, date, author, extractdPath string) error {
	opf := findOPF(extractdPath)
	if opf == "" {
		return fmt.Errorf("content.opf not found")
	}

	data, err := os.ReadFile(opf)
	if err != nil {
		return err
	}

	var pkg Package
	if err := xml.Unmarshal(data, &pkg); err != nil {
		return err
	}

	// Only update if field is empty
	if strings.TrimSpace(pkg.MetaData.Title) == "" && title != "" {
		pkg.MetaData.Title = title
	}
	if strings.TrimSpace(pkg.MetaData.Author) == "" && author != "" {
		pkg.MetaData.Author = author
	}
	if strings.TrimSpace(pkg.MetaData.Description) == "" && description != "" {
		pkg.MetaData.Description = description
	}
	if strings.TrimSpace(pkg.MetaData.Date) == "" && date != "" {
		pkg.MetaData.Date = date
	}

	// Marshal back to XML
	output, err := xml.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return err
	}

	// Add XML header
	output = []byte(xml.Header + string(output))

	return os.WriteFile(opf, output, 0644)
}

func Compress(dirPath string, outputEpub string) error {
	epubFile, err := os.Create(outputEpub)
	if err != nil {
		return err
	}
	defer epubFile.Close()

	zipWriter := zip.NewWriter(epubFile)
	defer zipWriter.Close()

	mimetypePath := filepath.Join(dirPath, "mimetype")
	mimetypeData, err := os.ReadFile(mimetypePath)
	if err != nil {
		return err
	}

	header := &zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store,
	}
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	if _, err := writer.Write(mimetypeData); err != nil {
		return err
	}

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() || strings.HasSuffix(path, "mimetype") {
			return nil
		}
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		writer, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}
