package service

import (
	"RSSFeed/models"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

func (svc *Service) addPathsToChan(paths []string, t models.PathType, wg *sync.WaitGroup) {
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
