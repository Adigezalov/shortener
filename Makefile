.PHONY: build build-with-version test benchmark profile clean fmt fmt-check lint check doc help deps staticlint staticlint-full run-with-trusted-subnet run-dev

# Установка зависимостей
deps:
	go mod download
	go mod tidy
	@echo "Installing development tools..."
	@if ! command -v goimports >/dev/null 2>&1; then \
		go install golang.org/x/tools/cmd/goimports@latest; \
	fi
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@echo "All dependencies installed successfully!"

# Сборка приложения
build:
	go build -o shortener ./cmd/shortener

# Сборка приложения с информацией о версии
build-with-version:
	go build -ldflags="-X 'main.buildVersion=$(VERSION)' -X 'main.buildDate=$(shell date +'%Y/%m/%d %H:%M:%S')' -X 'main.buildCommit=$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)'" -o shortener ./cmd/shortener

# Запуск тестов
test:
	go test ./...

# Запуск бенчмарков
benchmark:
	go test -bench=. -benchmem ./benchmarks/...

# Полный workflow профилирования
profile:
	./scripts/benchmark_workflow.sh

# Сбор профилей (требует запущенного приложения с профилированием)
collect-profiles:
	./scripts/collect_profiles.sh

# Сравнение профилей
compare-profiles:
	./scripts/compare_profiles.sh

# Форматирование кода
fmt:
	gofmt -w .
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	elif [ -f ~/go/bin/goimports ]; then \
		~/go/bin/goimports -w .; \
	elif [ -f $$(go env GOPATH)/bin/goimports ]; then \
		$$(go env GOPATH)/bin/goimports -w .; \
	else \
		echo "goimports not found, installing..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
		$$(go env GOPATH)/bin/goimports -w .; \
	fi

# Проверка форматирования (для CI/CD)
fmt-check:
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "The following files are not formatted:"; \
		gofmt -l .; \
		echo "Run 'make fmt' to fix formatting"; \
		exit 1; \
	fi
	@echo "All files are properly formatted"

# Линтинг кода
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, using go vet instead"; \
		go vet ./...; \
	fi

# Статический анализ с помощью multichecker (только production код)
staticlint:
	go run ./cmd/staticlint ./cmd/shortener ./internal/auth ./internal/config ./internal/database ./internal/handlers ./internal/logger ./internal/middleware ./internal/models ./internal/profiling ./internal/shortener ./internal/storage

# Запуск приложения с доверенной подсетью для внутренних эндпоинтов
run-with-trusted-subnet:
	TRUSTED_SUBNET=192.168.0.0/16 ./shortener

# Запуск приложения для локальной разработки с доверенной подсетью
run-dev:
	TRUSTED_SUBNET=127.0.0.1/32 ./shortener

# Полный статический анализ (включая тесты)
staticlint-full:
	go run ./cmd/staticlint ./cmd/... ./internal/...

# Комплексная проверка качества кода
check: fmt-check lint staticlint test
	@echo "All checks passed!"

# Генерация документации
doc:
	@echo "=== Документация пакетов ==="
	@echo ""
	@echo "Handlers:"
	@go doc -short ./internal/handlers
	@echo ""
	@echo "Models:"
	@go doc -short ./internal/models
	@echo ""
	@echo "Config:"
	@go doc -short ./internal/config
	@echo ""
	@echo "Shortener:"
	@go doc -short ./internal/shortener
	@echo ""
	@echo "=== Примеры использования ==="
	@go test -run "^Example$$" ./examples/
	@echo ""
	@echo "Для просмотра полной документации:"
	@echo "  godoc -http=:6060"
	@echo "  Затем откройте http://localhost:6060/pkg/github.com/Adigezalov/shortener/"

# Очистка профилей и бинарников
clean:
	rm -f shortener
	rm -f benchmarks/profiles/*.pprof

# Запуск приложения с профилированием
run-with-profiling:
	PROFILING_ENABLED=true PROFILING_PORT=:6060 ./shortener

# Справка
help:
	@echo "Доступные команды:"
	@echo "  deps               - Установить зависимости и инструменты разработки"
	@echo "  build              - Собрать приложение"
	@echo "  build-with-version - Собрать приложение с информацией о версии (VERSION=x.x.x make build-with-version)"
	@echo "  test               - Запустить тесты"
	@echo "  benchmark          - Запустить бенчмарки"
	@echo "  profile            - Полный workflow профилирования"
	@echo "  collect-profiles   - Собрать профили (требует запущенного приложения)"
	@echo "  compare-profiles   - Сравнить профили base.pprof и result.pprof"
	@echo "  run-with-profiling - Запустить приложение с профилированием"
	@echo "  fmt                - Форматировать код (gofmt + goimports)"
	@echo "  fmt-check          - Проверить форматирование (для CI/CD)"
	@echo "  lint               - Запустить линтер (golangci-lint или go vet)"
	@echo "  staticlint         - Запустить статический анализ (multichecker, только production код)"
	@echo "  staticlint-full    - Запустить полный статический анализ (multichecker, включая тесты)"
	@echo "  check              - Комплексная проверка (форматирование + линтинг + статический анализ + тесты)"
	@echo "  doc                - Генерация и просмотр документации"
	@echo "  clean              - Очистить профили и бинарники"
	@echo "  run-with-trusted-subnet - Запустить приложение с доверенной подсетью (192.168.0.0/16)"
	@echo "  run-dev            - Запустить для локальной разработки (127.0.0.1)"
	@echo "  help               - Показать эту справку"