package service

import (
	"RSSFeed/models"
	"RSSFeed/store"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
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
	//intiialisng 3 workers to process urls consurrently
	for i := 0; i < workersURL; i++ {
		go svc.RSSPathProcessor()
	}

	svc.wgFileWrite.Add(1)
	go svc.writeToHTMLFile()

	return svc
}

func (svc *Service) RequestProcessor(reqBody models.RequestBody) {
	defer svc.WgRequest.Done()
	wgAddPath := &sync.WaitGroup{}
	wgAddPath.Add(2)
	go svc.addPathsToChan(reqBody.FilePaths, models.TypeFile, wgAddPath)
	go svc.addPathsToChan(reqBody.Urls, models.TypeUrl, wgAddPath)
	wgAddPath.Wait()
}

func (svc *Service) RSSPathProcessor() {
	for path := range svc.paths {
		var rssData *models.RSS
		var err error

		if path.PathType == models.TypeFile {
			rssData, err = runWithRetry(svc.retryLimit, path.Path, ExtractRSSDataFromFile)
		} else {
			rssData, err = runWithRetry(svc.retryLimit, path.Path, FetchRSSDataFromURL)
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
	defer svc.wgFileWrite.Done()
	for rssItems := range svc.rssItemsChan {
		err := svc.file.WriteToFile(rssItems)
		if err != nil {
			slog.Info("file write successful")
		}
	}
}

func (svc *Service) parseItemAndAddToWrite(item models.Item, wg *sync.WaitGroup) {
	// adding recovery for waitgroup
	defer func() {
		if r := recover(); r != nil {
			slog.Error(fmt.Sprintf("Panic in 'parseItemAndAddToWrite'. Error:%v\n", r))
		}
	}()

	defer wg.Done()
	title, pubDate := item.Title, item.PubDate
	outputString := fmt.Sprintf("%v\t: %v", pubDate, title)
	svc.rssItemsChan <- outputString
}

func ExtractRSSDataFromFile(path string) (*models.RSS, error) {
	rssFile, err := os.Open(path)
	if err != nil {
		slog.Error(fmt.Sprintf("Error opening file %v:. Err:%v", path, err))
		return nil, err
	}
	defer rssFile.Close()
	return rssParseUtil(rssFile)
}

func FetchRSSDataFromURL(url string) (*models.RSS, error) {
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

func (svc *Service) Close() {
	svc.WgRequest.Wait()
	close(svc.paths)
	close(svc.rssItemsChan)
	svc.wgFileWrite.Wait()
}
