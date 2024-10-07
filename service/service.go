package service

import (
	"RSSFeed/models"
	"RSSFeed/store"
	"fmt"
	"log/slog"
	"sync"
)

type Service struct {
	file         *store.FileStore
	paths        chan models.Path
	rssItemsChan chan string
	retryLimit   int
	WgRequest    *sync.WaitGroup // this is to handle the number of requests running before shutting down
	wgFileWrite  *sync.WaitGroup // this is to handle the number of requests running before shutting down
}

func New(fs *store.FileStore) *Service {
	svc := &Service{
		paths: make(chan models.Path),
		// making channel buffered of size 100
		rssItemsChan: make(chan string, 100),
		file:         fs,
		retryLimit:   3,
		WgRequest:    &sync.WaitGroup{},
		wgFileWrite:  &sync.WaitGroup{},
	}

	workersURL := 3
	//intiialisng 3 workers to process urls concurrently
	for i := 0; i < workersURL; i++ {
		go svc.pathProcessor()
	}

	svc.wgFileWrite.Add(1)
	go svc.writeToHTMLFile()

	return svc
}

func (svc *Service) AddPathsToChan(reqBody models.RequestBody) {
	defer panicRecover("AddPathsToChan")
	defer svc.WgRequest.Done()
	wgAddPath := &sync.WaitGroup{}
	wgAddPath.Add(2)
	go svc.addPathsToChan(reqBody.FilePaths, models.TypeFile, wgAddPath)
	go svc.addPathsToChan(reqBody.Urls, models.TypeUrl, wgAddPath)
	wgAddPath.Wait()
}

func (svc *Service) pathProcessor() {
	defer panicRecover("RSSPathProcessor")

	for path := range svc.paths {
		var rssData *models.RSS
		var err error

		if path.PathType == models.TypeFile {
			rssData, err = runWithRetry(svc.retryLimit, path.Path, extractRSSDataFromFile)
		} else {
			rssData, err = runWithRetry(svc.retryLimit, path.Path, fetchRSSDataFromURL)
		}
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to fetch RSS Data from %v, Err:%v after retry. Dropping this path", path.Path, err))
			continue
		}

		svc.processRSSData(rssData)
	}
}

func (svc *Service) processRSSData(rssData *models.RSS) {
	//process each item in a separate go routine
	wg := &sync.WaitGroup{}
	wg.Add(len(rssData.Channel.Items))
	for _, item := range rssData.Channel.Items {
		go svc.parseItemAndAddToWrite(item, wg)
	}
	wg.Wait()
}

func (svc *Service) writeToHTMLFile() {
	defer panicRecover("writeToHTMLFile")

	defer svc.wgFileWrite.Done()
	for rssItems := range svc.rssItemsChan {
		err := svc.file.WriteToFile(rssItems)
		if err != nil {
			slog.Info("file write unsuccessful")
		}
	}
}

func (svc *Service) parseItemAndAddToWrite(item models.Item, wg *sync.WaitGroup) {
	defer panicRecover("parseItemAndAddToWrite")

	defer wg.Done()
	title, pubDate := item.Title, item.PubDate
	outputString := fmt.Sprintf("%v\t: %v", pubDate, title)
	svc.rssItemsChan <- outputString
}

func (svc *Service) Close() {
	svc.WgRequest.Wait()
	close(svc.paths)
	close(svc.rssItemsChan)
	svc.wgFileWrite.Wait()
}
