package main

import (
	"RSSFeed/handlers"
	"RSSFeed/service"
	"RSSFeed/store"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	outputFilePath := `C:\Users\Mohit\Desktop\RSSFeed\rssFeed.html`
	fileStore := store.New(outputFilePath)

	svc := service.New(fileStore)

	h := handlers.New(svc, outputFilePath)

	sm := mux.NewRouter()
	sm.HandleFunc("/rssPath", h.PathHandler)
	sm.HandleFunc("/rssFeeds", h.FeedHandler)

	port := "8000"
	server := http.Server{
		Handler: sm,
		Addr:    ":" + port,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			slog.Error("erro running server", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	slog.Info("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server unable to shutdown gracefully: %v", err)
	}

	svc.Close()
	slog.Info("Server stopped gracefully")
}
