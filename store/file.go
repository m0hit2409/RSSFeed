package store

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

const endHTMLContent = `</ul>
</body>
</html>`

const replaceHTMLContent = `</ul>`

type FileStore struct {
	path string
}

func New(path string) *FileStore {
	return &FileStore{
		path: path,
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
