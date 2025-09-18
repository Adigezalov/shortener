package utils

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProfileCollector(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	collector := NewProfileCollector(tempDir)

	t.Run("CollectHeapProfile", func(t *testing.T) {
		filename := "test_heap.pprof"
		err := collector.CollectHeapProfile(filename)
		if err != nil {
			t.Fatalf("Failed to collect heap profile: %v", err)
		}

		// Проверяем, что файл создан
		profilePath := filepath.Join(tempDir, filename)
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			t.Fatalf("Profile file was not created: %s", profilePath)
		}

		// Проверяем, что файл не пустой
		info, err := os.Stat(profilePath)
		if err != nil {
			t.Fatalf("Failed to stat profile file: %v", err)
		}
		if info.Size() == 0 {
			t.Fatalf("Profile file is empty")
		}
	})

	t.Run("CollectGoroutineProfile", func(t *testing.T) {
		filename := "test_goroutine.pprof"
		err := collector.CollectGoroutineProfile(filename)
		if err != nil {
			t.Fatalf("Failed to collect goroutine profile: %v", err)
		}

		// Проверяем, что файл создан
		profilePath := filepath.Join(tempDir, filename)
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			t.Fatalf("Profile file was not created: %s", profilePath)
		}
	})

	t.Run("CollectCPUProfile", func(t *testing.T) {
		filename := "test_cpu.pprof"
		// Используем короткую длительность для теста
		err := collector.CollectCPUProfile(filename, 100*time.Millisecond)
		if err != nil {
			t.Fatalf("Failed to collect CPU profile: %v", err)
		}

		// Проверяем, что файл создан
		profilePath := filepath.Join(tempDir, filename)
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			t.Fatalf("Profile file was not created: %s", profilePath)
		}
	})
}
