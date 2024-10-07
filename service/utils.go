package service

import (
	"RSSFeed/models"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

func (svc *Service) addPathsToChan(paths []string, t models.PathType, wg *sync.WaitGroup) {
	defer panicRecover("addPathsToChan")
	defer wg.Done()
	for _, path := range paths {
		svc.paths <- models.Path{PathType: t, Path: path}
	}
}

func runWithRetry(limit int, path string, f func(path string) (*models.RSS, error)) (*models.RSS, error) {
	backoff := 1
	for {
		res, err := f(path)
		if err == nil {
			return res, err
		}
		if limit == 0 {
			return res, err
		}
		slog.Error(fmt.Sprintf("Retrying fetch RSS Data from %v, Err:%v", path, err))
		time.Sleep(time.Second * time.Duration(backoff))
		backoff *= 2
		limit--
	}
}

func panicRecover(funcName string) {
	if panicErr := recover(); panicErr != nil {
		slog.Error(fmt.Sprintf("Recovered from panic in %v. Err:%v", funcName, panicErr))
	}
}

func extractRSSDataFromFile(path string) (*models.RSS, error) {
	rssFile, err := os.Open(path)
	if err != nil {
		slog.Error(fmt.Sprintf("Error opening file %v:. Err:%v", path, err))
		return nil, err
	}
	defer rssFile.Close()
	return rssParseUtil(rssFile)
}

func fetchRSSDataFromURL(url string) (*models.RSS, error) {
	response, err := http.Get(url)
	if err != nil {
		slog.Error(fmt.Sprintf("Error fetching dat from %v:. Err:%v", url, err))
		return nil, err
	}
	defer response.Body.Close()

	return rssParseUtil(response.Body)
}

func rssParseUtil(r io.Reader) (*models.RSS, error) {
	byteValue, err := io.ReadAll(r)
	if err != nil {
		slog.Error(fmt.Sprintf("Error reading byte. Err:%v", err))
		return nil, err
	}

	var rss models.RSS
	err = xml.Unmarshal(byteValue, &rss)
	if err != nil {
		slog.Error(fmt.Sprintf("Error unmarshalling XML. Err:%v", err))
		return nil, err
	}
	return &rss, nil
}
