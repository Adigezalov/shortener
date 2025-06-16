package handlers

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_PingDB(t *testing.T) {
	tests := []struct {
		name           string
		db             Pinger
		expectedStatus int
		prepareMock    func(p *MockPinger)
	}{
		{
			name:           "база_данных_не_настроена",
			db:             nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "база_данных_работает",
			db:             &MockPinger{},
			expectedStatus: http.StatusOK,
			prepareMock: func(p *MockPinger) {
				p.On("Ping").Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if mock, ok := tt.db.(*MockPinger); ok && tt.prepareMock != nil {
				tt.prepareMock(mock)
			}

			h := &Handler{
				db: tt.db,
			}

			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()

			h.PingDB(w, req)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.expectedStatus, result.StatusCode, "Expected status code %d, but got %d for test case '%s'", tt.expectedStatus, result.StatusCode, tt.name)
		})
	}
}
