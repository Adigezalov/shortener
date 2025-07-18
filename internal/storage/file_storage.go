package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/Adigezalov/shortener/internal/models"
)

// FileStorage реализует хранилище URL в файле
type FileStorage struct {
	urls       map[string]string   // id -> original_url
	urlToID    map[string]string   // original_url -> id
	userURLs   map[string][]string // userID -> []shortURL
	mu         sync.RWMutex
	filePath   string
	fileLock   *os.File
	flushQueue chan record
}

type record struct {
	ShortID     string `json:"short_id"`
	OriginalURL string `json:"original_url"`
}

// NewFileStorage создает новое файловое хранилище URL
func NewFileStorage(filePath string) *FileStorage {
	s := &FileStorage{
		urls:       make(map[string]string),
		urlToID:    make(map[string]string),
		userURLs:   make(map[string][]string),
		filePath:   filePath,
		flushQueue: make(chan record, 100),
	}

	// Создаем директорию, если её нет
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return s
	}

	// Пытаемся загрузить существующие данные
	if err := s.load(); err != nil {
		return s
	}

	// Запускаем горутину для асинхронной записи
	go s.flushWorker()

	return s
}

// Add добавляет новый URL в хранилище
func (s *FileStorage) Add(id string, url string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем, есть ли уже такой URL
	if existingID, found := s.urlToID[url]; found {
		return existingID, true, nil
	}

	// Добавляем новый URL
	s.urls[id] = url
	s.urlToID[url] = id

	// Отправляем на запись в файл
	s.flushQueue <- record{
		ShortID:     id,
		OriginalURL: url,
	}

	return id, false, nil
}

// AddWithUser добавляет новый URL в хранилище с привязкой к пользователю
func (s *FileStorage) AddWithUser(id string, url string, userID string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем, есть ли уже такой URL
	if existingID, found := s.urlToID[url]; found {
		return existingID, true, nil
	}

	// Добавляем новый URL
	s.urls[id] = url
	s.urlToID[url] = id

	// Добавляем URL к пользователю
	s.userURLs[userID] = append(s.userURLs[userID], id)

	// Отправляем на запись в файл
	s.flushQueue <- record{
		ShortID:     id,
		OriginalURL: url,
	}

	return id, false, nil
}

// Get возвращает оригинальный URL по идентификатору
func (s *FileStorage) Get(id string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.urls[id]
	return url, ok
}

// FindByOriginalURL ищет ID по оригинальному URL
func (s *FileStorage) FindByOriginalURL(url string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.urlToID[url]
	return id, ok
}

// GetUserURLs возвращает все URL пользователя
func (s *FileStorage) GetUserURLs(userID string) ([]models.UserURL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	shortURLs, exists := s.userURLs[userID]
	if !exists || len(shortURLs) == 0 {
		return []models.UserURL{}, nil
	}

	result := make([]models.UserURL, 0, len(shortURLs))
	for _, shortURL := range shortURLs {
		if originalURL, exists := s.urls[shortURL]; exists {
			result = append(result, models.UserURL{
				ShortURL:    shortURL,
				OriginalURL: originalURL,
			})
		}
	}

	return result, nil
}

// Close закрывает хранилище и освобождает ресурсы
func (s *FileStorage) Close() error {
	close(s.flushQueue)
	if s.fileLock != nil {
		if err := syscall.Flock(int(s.fileLock.Fd()), syscall.LOCK_UN); err != nil {
			return fmt.Errorf("ошибка снятия блокировки файла: %w", err)
		}
		if err := s.fileLock.Close(); err != nil {
			return fmt.Errorf("ошибка закрытия файла блокировки: %w", err)
		}
		if err := os.Remove(s.fileLock.Name()); err != nil {
			return fmt.Errorf("ошибка удаления файла блокировки: %w", err)
		}
	}
	return nil
}

// load загружает данные из файла
func (s *FileStorage) load() error {
	file, err := os.OpenFile(s.filePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var r record
		if err := json.Unmarshal(scanner.Bytes(), &r); err != nil {
			continue
		}
		s.urls[r.ShortID] = r.OriginalURL
		s.urlToID[r.OriginalURL] = r.ShortID
	}

	return scanner.Err()
}

// flushWorker асинхронно записывает URL в файл
func (s *FileStorage) flushWorker() {
	file, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for r := range s.flushQueue {
		data, err := json.Marshal(r)
		if err != nil {
			continue
		}
		writer.Write(data)
		writer.WriteString("\n")
		writer.Flush()
	}
}
