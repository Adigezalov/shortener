package service

import "github.com/Adigezalov/shortener/internal/storage"

type URLServiceImpl struct {
	storage  storage.URLReadWriter
	baseURL  string
	filePath string
}
