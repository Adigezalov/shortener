package config

import (
	"flag"
	"os"
	"reflect"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		env            map[string]string
		wantServerAddr string
		wantBaseURL    string
	}{
		{
			name:           "default_values",
			args:           []string{},
			env:            map[string]string{},
			wantServerAddr: "localhost:8080",
			wantBaseURL:    "http://localhost:8080",
		},
		{
			name:           "flags_override_defaults",
			args:           []string{"-a=:9090", "-b=http://example.com"},
			env:            map[string]string{},
			wantServerAddr: ":9090",
			wantBaseURL:    "http://example.com",
		},
		{
			name: "env_overrides_defaults",
			args: []string{},
			env: map[string]string{
				"SERVER_ADDRESS": ":7070",
				"BASE_URL":       "http://env.example.com",
			},
			wantServerAddr: ":7070",
			wantBaseURL:    "http://env.example.com",
		},
		{
			name: "flags_override_env",
			args: []string{"-a=:6060", "-b=http://flag.example.com"},
			env: map[string]string{
				"SERVER_ADDRESS": ":8081",
				"BASE_URL":       "http://env.example.com",
			},
			wantServerAddr: ":6060",
			wantBaseURL:    "http://flag.example.com",
		},
		{
			name:           "trailing_slash_in_base_URL",
			args:           []string{"-b=http://example.com/"},
			env:            map[string]string{},
			wantServerAddr: "localhost:8080",
			wantBaseURL:    "http://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Сохраняем оригинальные аргументы и окружение
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = append([]string{"cmd"}, tt.args...)

			oldEnv := map[string]string{}
			for k, v := range tt.env {
				oldEnv[k] = os.Getenv(k)
				os.Setenv(k, v)
			}
			defer func() {
				for k, v := range oldEnv {
					os.Setenv(k, v)
				}
			}()

			// Сбрасываем флаги перед каждым тестом
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			got := ParseFlags()

			if got.ServerAddress != tt.wantServerAddr {
				t.Errorf("ParseFlags() ServerAddress = %v, want %v", got.ServerAddress, tt.wantServerAddr)
			}
			if got.BaseURL != tt.wantBaseURL {
				t.Errorf("ParseFlags() BaseURL = %v, want %v", got.BaseURL, tt.wantBaseURL)
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	// Проверяем, что значения по умолчанию соответствуют ожидаемым
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd"}

	// Очищаем окружение для этого теста
	oldServerAddr := os.Getenv("SERVER_ADDRESS")
	oldBaseURL := os.Getenv("BASE_URL")
	oldFileStoragePath := os.Getenv("FILE_STORAGE_PATH")

	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
	os.Unsetenv("FILE_STORAGE_PATH")

	defer func() {
		os.Setenv("SERVER_ADDRESS", oldServerAddr)
		os.Setenv("BASE_URL", oldBaseURL)
		os.Setenv("FILE_STORAGE_PATH", oldFileStoragePath)
	}()

	// Сбрасываем флаги
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	cfg := ParseFlags()

	defaults := &Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "storage.json",
	}

	if !reflect.DeepEqual(cfg, defaults) {
		t.Errorf("Default config = %+v, want %+v", cfg, defaults)
	}
}
