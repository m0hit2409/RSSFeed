package models

type RequestBody struct {
	Urls      []string `json:"urls"`
	FilePaths []string `json:"filePaths"`
}
