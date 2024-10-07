package handlers

import (
	"RSSFeed/models"
	"RSSFeed/service"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
)

type Handler struct {
	svc      *service.Service
	filePath string
}

func New(svc *service.Service, filePath string) *Handler {
	return &Handler{
		svc:      svc,
		filePath: filePath,
	}
}

func (h *Handler) PathHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Recovered from panic in HTTP handler: %v", r)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}()
	if r.Method != http.MethodPost {
		slog.Error(fmt.Sprintf("Invalid request method:%v  in HTTP handler: %v", r.Method, r.URL))
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		slog.Error("Unable to read request body")
		return
	}
	defer r.Body.Close()

	var requestBody models.RequestBody
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		http.Error(w, "Unsupported request body format", http.StatusBadRequest)
		slog.Error("Unsupported request body format")
		return
	}

	response := map[string]string{"status": "request acknowledged"}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Unable to create response", http.StatusInternalServerError)
		slog.Error("Unable to create response for request")
		return
	}

	h.svc.WgRequest.Add(1)
	go h.svc.RequestProcessor(requestBody)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(jsonResponse)
	slog.Info(fmt.Sprintf("Request successful '%v|%v'", r.Method, r.URL))
}

func (h *Handler) FeedHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Recovered from panic in HTTP handler: %v", r)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}()
	file, err := os.ReadFile(h.filePath)
	if err != nil {
		http.Error(w, "Some error occurred", http.StatusInternalServerError)
		slog.Error(fmt.Sprintf("Could not read file '%v'. Err:%v", h.filePath, err))
		return
	}

	w.Header().Set("Content-Type", "text/html")

	_, err = fmt.Fprint(w, string(file))
	if err != nil {
		http.Error(w, "Unable to generate response", http.StatusInternalServerError)
		slog.Error(fmt.Sprintf("Could not write html response. Err:%v", err))
		return
	}

	slog.Info(fmt.Sprintf("Request successful '%v|%v'", r.Method, r.URL))
}
