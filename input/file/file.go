package file

import (
	"os"
	"bufio"

	"log"
)

type FileInput struct {
	fileInput string
	f *os.File
	scanner *bufio.Scanner
	docCh chan map[string]interface{}
}

func NewFileInput(filePath string, docCh chan map[string]interface{}) (*FileInput) {
	
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(f)
	fileInput := &FileInput {
		f: f,
		scanner: scanner,
		docCh: docCh,
	}
	return fileInput
}

func (input *FileInput) Run() {
	defer input.f.Close()
	
	for input.scanner.Scan() {
		input.docCh <- map[string]interface{}{"message": input.scanner.Text()}
	}

	if err := input.scanner.Err(); err != nil {
		log.Fatalf("Failed to read from file %s", err)
	}
}