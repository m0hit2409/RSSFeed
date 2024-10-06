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

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		slog.Error(fmt.Sprintf("Invalid request method:%v for '%v'", r.Method, r.URL))
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

	go h.svc.RequestProcessor(requestBody)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(jsonResponse)
	slog.Info(fmt.Sprintf("Request successful '%v|%v'", r.Method, r.URL))
}

func (h *Handler) FeedHandler(w http.ResponseWriter, r *http.Request) {
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
