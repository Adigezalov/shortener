package shortener

import "testing"

// TestShorten_GeneratesUniqueIDs проверяет, что Shorten возвращает строки нужной длины и разные значения
func TestShorten_GeneratesUniqueIDs(t *testing.T) {
	service := New("http://localhost:8080")
	seen := make(map[string]bool)

	for i := 0; i < 1000; i++ {
		id := service.Shorten("https://example.com")
		if len(id) != 8 {
			t.Errorf("expected id length 8, got %d", len(id))
		}
		if seen[id] {
			t.Errorf("duplicate id generated: %s", id)
		}
		seen[id] = true
	}
}

// TestBuildShortURL проверяет корректность построения короткого URL
func TestBuildShortURL(t *testing.T) {
	baseURL := "http://localhost:8080"
	service := New(baseURL)

	id := "abc12345"
	expected := baseURL + "/" + id
	result := service.BuildShortURL(id)

	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}
