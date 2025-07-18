package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Adigezalov/shortener/internal/database"
	"github.com/Adigezalov/shortener/internal/logger"
	"github.com/Adigezalov/shortener/internal/models"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"sync"
	"syscall"
)

// MemoryStorage реализует хранилище URL с опциональным сохранением в файл
type MemoryStorage struct {
	urls     map[string]string   // id -> original_url
	urlToID  map[string]string   // original_url -> id (обратный индекс)
	userURLs map[string][]string // userID -> []shortURL (URL пользователя)
	mu       sync.RWMutex        // мьютекс для защиты данных
	nextID   int                 // счетчик ID для новых записей

	// Поля для работы с файлом (используются только если storagePath не пустой)
	storagePath string                // путь к файлу хранения
	flushQueue  chan models.URLRecord // канал для асинхронной записи
	fileLock    *os.File              // файловый дескриптор для блокировки
	batchSize   int                   // размер пакета для записи
	batchBuffer []models.URLRecord    // буфер для пакетной записи
	batchMu     sync.Mutex            // мьютекс для буфера
	fileMode    bool                  // флаг работы с файлом
}

// NewMemoryStorage создает новое хранилище URL
// Если путь к файлу не пустой, данные будут сохраняться в файл
// и восстанавливаться из него при запуске
func NewMemoryStorage(storagePath string) *MemoryStorage {
	storage := &MemoryStorage{
		urls:        make(map[string]string),
		urlToID:     make(map[string]string),
		userURLs:    make(map[string][]string),
		nextID:      1,
		storagePath: storagePath,
		fileMode:    storagePath != "",
	}

	// Если указан путь к файлу, инициализируем файловое хранилище
	if storage.fileMode {
		storage.flushQueue = make(chan models.URLRecord, 100)
		storage.batchSize = 1
		storage.batchBuffer = make([]models.URLRecord, 0, 10)

		// Создаем блокировку файла
		if err := storage.acquireLock(); err != nil {
			logger.Logger.Error("Ошибка блокировки файла хранения", zap.Error(err))
		}

		// Восстанавливаем данные из файла
		if err := storage.restore(); err != nil {
			logger.Logger.Error("Ошибка восстановления данных", zap.Error(err))
		}

		// Запускаем горутину для асинхронной записи
		go storage.flushWorker()

		logger.Logger.Info("Создано хранилище URL с сохранением в файл",
			zap.String("path", storagePath))
	} else {
		logger.Logger.Info("Создано хранилище URL в памяти")
	}

	return storage
}

// Add добавляет новый URL в хранилище или возвращает существующий ID
func (s *MemoryStorage) Add(id string, url string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем, есть ли уже такой URL
	if existingID, found := s.urlToID[url]; found {
		return existingID, true, database.ErrURLConflict
	}

	// Добавляем новый URL
	s.urls[id] = url
	s.urlToID[url] = id

	// Если включен режим файла, добавляем запись в очередь на сохранение
	if s.fileMode {
		uuid := fmt.Sprintf("%d", s.nextID)
		s.flushQueue <- models.URLRecord{
			UUID:        uuid,
			ShortURL:    id,
			OriginalURL: url,
		}
	}

	s.nextID++
	return id, false, nil
}

// Get возвращает оригинальный URL по идентификатору
func (s *MemoryStorage) Get(id string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.urls[id]
	return url, ok
}

// FindByOriginalURL ищет ID по оригинальному URL
func (s *MemoryStorage) FindByOriginalURL(url string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.urlToID[url]
	return id, ok
}

// AddWithUser добавляет новый URL в хранилище с привязкой к пользователю
func (s *MemoryStorage) AddWithUser(id string, url string, userID string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем, есть ли уже такой URL
	if existingID, found := s.urlToID[url]; found {
		return existingID, true, database.ErrURLConflict
	}

	// Добавляем новый URL
	s.urls[id] = url
	s.urlToID[url] = id

	// Добавляем URL к пользователю
	s.userURLs[userID] = append(s.userURLs[userID], id)

	// Если включен режим файла, добавляем запись в очередь на сохранение
	if s.fileMode {
		uuid := fmt.Sprintf("%d", s.nextID)
		s.flushQueue <- models.URLRecord{
			UUID:        uuid,
			ShortURL:    id,
			OriginalURL: url,
		}
	}

	s.nextID++
	return id, false, nil
}

// GetUserURLs возвращает все URL пользователя
func (s *MemoryStorage) GetUserURLs(userID string) ([]models.UserURL, error) {
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

// acquireLock блокирует файл хранения
func (s *MemoryStorage) acquireLock() error {
	// Создаем файл блокировки
	lockFile := s.storagePath + ".lock"

	// Создаем директорию, если она не существует
	dir := filepath.Dir(lockFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("ошибка создания директории: %w", err)
	}

	// Открываем файл блокировки
	var err error
	s.fileLock, err = os.OpenFile(lockFile, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла блокировки: %w", err)
	}

	// Пытаемся установить блокировку
	err = syscall.Flock(int(s.fileLock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		s.fileLock.Close()
		return fmt.Errorf("не удалось получить блокировку файла (возможно, файл уже используется): %w", err)
	}

	logger.Logger.Info("Файл хранения успешно заблокирован", zap.String("lock_file", lockFile))
	return nil
}

// restore восстанавливает данные из файла
func (s *MemoryStorage) restore() error {
	// Проверяем существование файла
	if _, err := os.Stat(s.storagePath); os.IsNotExist(err) {
		logger.Logger.Info("Файл хранения не найден, создаем новое хранилище",
			zap.String("path", s.storagePath))
		return nil
	}

	// Открываем файл
	file, err := os.Open(s.storagePath)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла хранения: %w", err)
	}
	defer file.Close()

	// Создаем сканер для чтения файла построчно
	scanner := bufio.NewScanner(file)
	maxID := 0

	// Читаем и обрабатываем каждую строку
	for scanner.Scan() {
		line := scanner.Text()
		var record models.URLRecord

		// Декодируем JSON
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			logger.Logger.Error("Ошибка декодирования записи URL",
				zap.String("line", line),
				zap.Error(err))
			continue
		}

		// Сохраняем URL в памяти
		s.urls[record.ShortURL] = record.OriginalURL
		s.urlToID[record.OriginalURL] = record.ShortURL

		// Обновляем счетчик ID
		if id := parseID(record.UUID); id > maxID {
			maxID = id
		}
	}

	// Проверяем ошибки сканирования
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("ошибка чтения файла хранения: %w", err)
	}

	// Устанавливаем следующий ID
	s.nextID = maxID + 1

	logger.Logger.Info("Данные успешно восстановлены из файла",
		zap.Int("records", len(s.urls)),
		zap.String("path", s.storagePath))

	return nil
}

// flushWorker асинхронно записывает URL в файл
func (s *MemoryStorage) flushWorker() {
	for record := range s.flushQueue {
		s.batchMu.Lock()
		// Добавляем запись в буфер
		s.batchBuffer = append(s.batchBuffer, record)

		// Если буфер заполнен, записываем пакет
		if len(s.batchBuffer) >= s.batchSize {
			if err := s.writeBatch(s.batchBuffer); err != nil {
				logger.Logger.Error("Ошибка записи пакета URL в файл", zap.Error(err))
			}
			// Очищаем буфер
			s.batchBuffer = make([]models.URLRecord, 0, s.batchSize)
		}
		s.batchMu.Unlock()
	}

	// Записываем оставшиеся записи при завершении
	s.batchMu.Lock()
	if len(s.batchBuffer) > 0 {
		if err := s.writeBatch(s.batchBuffer); err != nil {
			logger.Logger.Error("Ошибка записи оставшихся URL в файл", zap.Error(err))
		}
	}
	s.batchMu.Unlock()
}

// writeBatch записывает пакет записей в файл
func (s *MemoryStorage) writeBatch(records []models.URLRecord) error {
	// Проверяем, есть ли что записывать
	if len(records) == 0 {
		return nil
	}

	// Создаем директорию, если она не существует
	dir := filepath.Dir(s.storagePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("ошибка создания директории: %w", err)
	}

	// Открываем файл для добавления
	file, err := os.OpenFile(s.storagePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

	// Создаем буферизованный писатель
	writer := bufio.NewWriter(file)

	// Записываем каждую запись
	for _, record := range records {
		data, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("ошибка кодирования записи: %w", err)
		}

		if _, err := writer.Write(data); err != nil {
			return fmt.Errorf("ошибка записи данных: %w", err)
		}

		if _, err := writer.Write([]byte("\n")); err != nil {
			return fmt.Errorf("ошибка записи переноса строки: %w", err)
		}
	}

	// Сбрасываем буфер на диск
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("ошибка сброса буфера: %w", err)
	}

	logger.Logger.Info("Записан пакет URL в файл",
		zap.Int("count", len(records)),
		zap.String("path", s.storagePath))

	return nil
}

// Close закрывает хранилище и освобождает ресурсы
func (s *MemoryStorage) Close() error {
	// Если работаем с файлом, закрываем файловые ресурсы
	if s.fileMode {
		// Закрываем канал flush
		close(s.flushQueue)

		// Снимаем блокировку файла
		if s.fileLock != nil {
			if err := syscall.Flock(int(s.fileLock.Fd()), syscall.LOCK_UN); err != nil {
				logger.Logger.Error("Ошибка снятия блокировки файла", zap.Error(err))
			}
			if err := s.fileLock.Close(); err != nil {
				logger.Logger.Error("Ошибка закрытия файла блокировки", zap.Error(err))
			}
			if err := os.Remove(s.fileLock.Name()); err != nil {
				logger.Logger.Error("Ошибка удаления файла блокировки", zap.Error(err))
			}
		}
	}

	return nil
}

// parseID преобразует строковый ID в int
func parseID(id string) int {
	var result int
	_, _ = fmt.Sscanf(id, "%d", &result)
	return result
}
