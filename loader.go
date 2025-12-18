// loader.go
package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type FileMetadata struct {
	path          string
	name          string
	size          string
	length        int
	newlineType   string
	encodingType  string
	extensionType string
	modifiedDate  string
}

type FileDetails struct {
	Content  string
	Metadata FileMetadata
}

func (m *FileMetadata) detectFileEncodingType(filePath string) {
	// feat: if --force-utf is true, then do this as usual
	// otherwise use bom to switch to other encoding types

	if m.encodingType != "" {
		return
	}

	// 1. Read the first chunk of the file
	// 		if contains null bytes or non-printable chars, abort
	//		(if chars are outside the printable ascii range)
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("could not open file to count lines: %v", err)
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	buf := make([]byte, 32*1024)

	// n = 1 big block
	n, err := reader.Read(buf)
	if err != nil {
		log.Fatalf("could not read file to count lines: %v", err)
		return
	}

	var chunk []byte
	if n > 0 {
		chunk = buf[:n]
	}

	for _, r := range string(chunk) {
		if r < 32 && r != 9 && r != 10 && r != 13 {
			log.Printf("file encoding detection aborted: non-printable character found: %v", r)
			return
		}
	}

	// 2. Check the BOM for encoding
	//		Have this enabled in a config file in the future

	// Read the bom from the first chunk
	if bytes.HasPrefix(chunk, []byte{0xEF, 0xBB, 0xBF}) {
		m.encodingType = "UTF-8"
	} else if bytes.HasPrefix(chunk, []byte{0xFF, 0xFE}) {
		m.encodingType = "UTF-16LE"
	} else if bytes.HasPrefix(chunk, []byte{0xFE, 0xFF}) {
		m.encodingType = "UTF-16BE"
	} else {
		m.encodingType = "UTF-8" // Default to UTF-8
	}

}

func (m *FileMetadata) detectFileNewlineType(filePath string) {
	if m.encodingType == "" {
		// Ensure that encoding type is valid first
		m.detectFileEncodingType(filePath)
	}

	// Read end of first line
	// switch on line ending characters
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("could not open file to detect newline type: %v", err)
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	line, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("could not read line to detect newline type: %v", err)
		return
	}

	if len(line) >= 2 && line[len(line)-2] == '\r' {
		m.newlineType = "CRLF"
	} else if len(line) >= 1 && line[len(line)-1] == '\r' {
		m.newlineType = "CR"
	} else {
		m.newlineType = "LF"
	}

}

func (m *FileMetadata) countFileLines(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("could not open file to count lines: %v", err)
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	buf := make([]byte, 32*1024)
	lineCount := 0
	readAny := false
	var lastByte byte

	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("could not read file to count lines: %v", err)
			return
		}

		if n > 0 {
			readAny = true
			chunk := buf[:n]
			lineCount += bytes.Count(chunk, []byte{'\n'})
			lastByte = chunk[n-1]
		}
	}

	if readAny && lastByte != '\n' {
		lineCount++
	}

	m.length = lineCount
}

func loadFileMetadata(filePath string) FileMetadata {
	// Details I want to load:
	// - filename
	// - file size, bytes
	// - file length, lines
	// - newline charater type (CRLF, LF, CR)
	// - encoding (utf-8, utf-16, etc)
	// - file extension/type
	// - modification datetime

	var metadata FileMetadata

	if filePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("could not get cwd: %v", err)
			return metadata
		}
		filePath = cwd
	}

	ext := filepath.Ext(filePath)

	info, err := os.Stat(filePath)
	if err != nil {
		log.Fatalf("could not stat file: %v", err)
		return metadata
	}

	// convert info.Size() from bytecount to KB if applicable
	fileSize := ""
	if info.Size() < 1024 {
		fileSize = strconv.FormatInt(info.Size(), 10) + "B"
	} else {
		fileSize = strconv.FormatInt(info.Size()/1024, 10) + "KB"
	}

	metadata.path = filePath
	metadata.name = info.Name()
	metadata.size = fileSize
	metadata.detectFileEncodingType(filePath)
	metadata.detectFileNewlineType(filePath)
	metadata.countFileLines(filePath)
	metadata.extensionType = ext
	metadata.modifiedDate = strings.Split(info.ModTime().String(), " ")[0] // Remove Seconds, only date

	return metadata
}

func loadFileContents(filePath string) string {
	if filePath == "" {
		return ""
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		// feat:
		// Open a new file
		panic(err)
	}
	return string(data)
}

// feat:
// load large file details using bufio.scanner + streaming?
// I have a feeling that searching wouldn't work with that, so must find another way
// ideas include indexing the file separately, or using a token database??
func loadFileDetails(filePath string) FileDetails {
	return FileDetails{
		Content:  loadFileContents(filePath),
		Metadata: loadFileMetadata(filePath),
	}
}
