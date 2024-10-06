package main

import (
	"RSSFeed/handlers"
	"RSSFeed/service"
	"RSSFeed/store"
	"fmt"
	"log"
	"net/http"
)

func main() {
	outputFilePath := `C:\Users\Mohit\Desktop\RSSFeed\rssFeed.html`
	fileStore := store.New(outputFilePath)

	svc := service.New(fileStore)

	h := handlers.New(svc, outputFilePath)

	http.HandleFunc("/rssPath", h.PathHandler)
	http.HandleFunc("/rssFeeds", h.FeedHandler)

	port := "8000"
	// Start the server on port 8080
	fmt.Println("Server listening on port:", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %s", err)
	}

}
