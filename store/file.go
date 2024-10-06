package store

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type FileStore struct {
	path string
}

func New(path string) *FileStore {
	return &FileStore{
		path: path,
	}
}

func (fs *FileStore) WriteToFile(itemSummary string) error {
	content, err := os.ReadFile(fs.path)
	if err != nil {
		slog.Error(fmt.Sprintf("Could not read file:%v. Err:%v", fs.path, err))
		return err
	}

	htmlContent := string(content)
	itemSummary = fmt.Sprintf("<li>%s</li>\n", itemSummary)

	updatedHTMLContent := strings.Replace(htmlContent, "</ul>", itemSummary+"</ul>", 1)

	err = os.WriteFile(fs.path, []byte(updatedHTMLContent), 0644)
	if err != nil {
		slog.Error(fmt.Sprintf("Could not update file:%v. Err:%v", fs.path, err))
		return err
	}

	slog.Info("New itme added in the file")
	return nil
}
