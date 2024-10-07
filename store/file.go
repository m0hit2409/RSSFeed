package store

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
)

type FileStore struct {
	path string
}

func New(path string) *FileStore {
	createFileIfNotExist(path)
	return &FileStore{
		path: path,
	}
}

func createFileIfNotExist(path string) {
	_, err := os.Stat(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("Some error occurred while reading content file. Err:%v", err)
	}
	if err != nil && errors.Is(err, os.ErrNotExist) {
		file, errFile := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_RDWR, 0644)
		if errFile != nil {
			log.Fatalf("Some error occurred while creating content file. Err:%v", err)
		}
		errWrite := appendFile(file, newTemplateContent)
		if errWrite != nil {
			log.Fatalf("Some error occurred while writing content file. Err:%v", err)
		}
	}
}

func (fs *FileStore) WriteToFile(itemSummary string) error {
	file, err := os.OpenFile(fs.path, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	// removing the ending html contents from file
	err = truncateFileAfterMatching(file, replaceHTMLContent)
	if err != nil {
		slog.Error("Error occurred in truncating file", "Err", err)
		return err
	}

	// creating new line to append
	appendContent := fmt.Sprintf("<li>%s</li>\n%s", itemSummary, endHTMLContent)

	err = appendFile(file, appendContent)
	if err != nil {
		slog.Error("Error occurred in appending file", "Err", err)
		return err
	}
	return nil
}

func appendFile(file *os.File, line string) error {
	_, err := file.WriteString(line)
	return err
}

func truncateFileAfterMatching(file *os.File, match string) error {
	scanner := bufio.NewScanner(file)

	var lastMatchOffset int64 = -1
	var currentOffset int64 = 0

	for scanner.Scan() {
		line := scanner.Text()

		lineLength := int64(len(line) + len("\n"))

		if strings.Contains(line, match) {
			lastMatchOffset = currentOffset
			break
		}

		currentOffset += lineLength
	}

	if lastMatchOffset == -1 {
		return fmt.Errorf("corrupt file")
	}

	return file.Truncate(lastMatchOffset)
}
