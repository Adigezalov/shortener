package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// SaveToFile сохраняет список записей в указанный файл
func SaveToFile(filePath string, records []URLRecord) error {
	// Нормализуем путь (особенно важно для Windows)
	filePath = filepath.Clean(filePath)

	// Создаем все необходимые директории
	dir := filepath.Dir(filePath)
	if dir != "." && dir != string(filepath.Separator) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, record := range records {
		line, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("json marshal error: %w", err)
		}
		if _, err := writer.Write(append(line, '\n')); err != nil {
			return fmt.Errorf("write error: %w", err)
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flush error: %w", err)
	}
	return nil
}

// LoadFromFile загружает список записей из указанного файла
func LoadFromFile(filePath string) ([]URLRecord, error) {
	var records []URLRecord

	file, err := os.Open(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return records, nil // файл ещё не создан — вернём пустой список
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record URLRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, scanner.Err()
}
