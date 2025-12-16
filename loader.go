// loader.go
package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
)

type FileMetadata struct {
	path             string
	name             string
	size             int64
	length           int
	newlineType      string
	encodingType     string
	extensionType    string
	modifiedDatetime string
}

type FileDetails struct {
	Content  string
	Metadata FileMetadata
}

func isValidUTF8(b []byte) bool {
	for i := 0; i < len(b); {
		if b[i] < 0x80 {
			i++
		} else if b[i] < 0xC0 {
			return false
		} else if b[i] < 0xE0 {
			if i+1 >= len(b) {
				return false
			}
			i += 2
		} else if b[i] < 0xF0 {
			if i+2 >= len(b) {
				return false
			}
			i += 3
		} else if b[i] < 0xF8 {
			if i+3 >= len(b) {
				return false
			}
			i += 4
		} else {
			return false
		}
	}
	return true
}

func (m *FileMetadata) detectFileEncodingType(filePath string) {
	// feat: if --force-utf8 is true, then do this as usual
	// otherwise try to do some bom magic mayb
	// rip chinese characters

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

	// if bytes contains binary values outside the ascii printable range, fail
	// the acceptable range is 32-126, 9 (tab), 10 (lf), 13 (cr)
	if !isValidUTF8(chunk) {
		log.Printf("file encoding detection aborted: invalid UTF-8 sequence found")
		return
	}

	for _, r := range string(chunk) {
		if r < 32 && r != 9 && r != 10 && r != 13 {
			log.Printf("file encoding detection aborted: non-printable character found: %v", r)
			return
		}
	}

	// 2. Check the BOM for encoding
	//		Have this enabled in a config file in the future

	// Default to utf-8
	// m.encodingType = "utf-8"
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
	//tab sizs if applicable xdd

	metadata.path = filePath
	metadata.name = info.Name()
	metadata.size = info.Size()
	metadata.detectFileEncodingType(filePath)
	metadata.detectFileNewlineType(filePath)
	metadata.countFileLines(filePath)
	metadata.extensionType = ext
	metadata.modifiedDatetime = info.ModTime().String()

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
