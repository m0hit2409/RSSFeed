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
	outputFilePath := `C:\Users\Mohit\Desktop\RSSFeed\output\rss_summary.html`
	fileStore := store.New(outputFilePath)

	svc := service.New(fileStore)

	h := handlers.New(svc, outputFilePath)

	sm := mux.NewRouter()
	sm.HandleFunc("/rss-path", h.PathHandler)
	sm.HandleFunc("/rss-feeds", h.FeedHandler)

	port := "8000"
	server := http.Server{
		Handler: sm,
		Addr:    ":" + port,
	}
	go func() {
		slog.Info("Starting server on", "port", port)
		if err := server.ListenAndServe(); err != nil {
			slog.Error("error running server", "Err", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	slog.Info("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server unable to shutdown gracefully: %v", "Err", err)
	}

	svc.Close()
	slog.Info("Server stopped gracefully")
}
