package storage

import "testing"

func TestNewMemoryStorage(t *testing.T) {
	storage := NewMemoryStorage("storage")

	if storage == nil {
		t.Fatal("expected non-nil MemoryStorage")
	}

	if storage.urlStore == nil {
		t.Error("expected urlStore to be initialized")
	}

	if storage.reverseStore == nil {
		t.Error("expected reverseStore to be initialized")
	}

	if len(storage.urlStore) != 0 {
		t.Errorf("expected urlStore to be empty, got %d items", len(storage.urlStore))
	}

	if len(storage.reverseStore) != 0 {
		t.Errorf("expected reverseStore to be empty, got %d items", len(storage.reverseStore))
	}
}

func TestMemoryStorage_Save(t *testing.T) {
	storage := NewMemoryStorage("tmp/storage")

	shortID := "abc123"
	originalURL := "https://example.com"

	err := storage.Save(shortID, originalURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gotURL, ok := storage.urlStore[shortID]
	if !ok {
		t.Errorf("shortID %q not found in urlStore", shortID)
	}
	if gotURL != originalURL {
		t.Errorf("expected originalURL %q, got %q", originalURL, gotURL)
	}

	gotShortID, ok := storage.reverseStore[originalURL]
	if !ok {
		t.Errorf("originalURL %q not found in reverseStore", originalURL)
	}
	if gotShortID != shortID {
		t.Errorf("expected shortID %q, got %q", shortID, gotShortID)
	}
}

func TestMemoryStorage_Get(t *testing.T) {
	storage := NewMemoryStorage("tmp/storage")

	shortID := "abc123"
	originalURL := "https://example.com"

	// Пустое хранилище — ключа нет
	_, exists := storage.Get(shortID)
	if exists {
		t.Errorf("expected shortID %q to not exist", shortID)
	}

	// Сохраняем и проверяем
	err := storage.Save(shortID, originalURL)
	if err != nil {
		t.Fatalf("unexpected error on save: %v", err)
	}

	gotURL, exists := storage.Get(shortID)
	if !exists {
		t.Errorf("expected shortID %q to exist after save", shortID)
	}
	if gotURL != originalURL {
		t.Errorf("expected originalURL %q, got %q", originalURL, gotURL)
	}
}

func TestMemoryStorage_Exists(t *testing.T) {
	storage := NewMemoryStorage("tmp/storage")

	shortID := "abc123"
	originalURL := "https://example.com"

	// До сохранения: URL не существует
	_, exists := storage.Exists(originalURL)
	if exists {
		t.Errorf("expected originalURL %q to not exist", originalURL)
	}

	// Сохраняем пару и проверяем
	err := storage.Save(shortID, originalURL)
	if err != nil {
		t.Fatalf("unexpected error on save: %v", err)
	}

	gotShortID, exists := storage.Exists(originalURL)
	if !exists {
		t.Errorf("expected originalURL %q to exist after save", originalURL)
	}
	if gotShortID != shortID {
		t.Errorf("expected shortID %q, got %q", shortID, gotShortID)
	}
}
