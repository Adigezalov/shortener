.PHONY: build test benchmark profile clean fmt fmt-check lint check help

# Сборка приложения
build:
	go build -o shortener ./cmd/shortener

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

# Комплексная проверка качества кода
check: fmt-check lint test
	@echo "All checks passed!"

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
	@echo "  build              - Собрать приложение"
	@echo "  test               - Запустить тесты"
	@echo "  benchmark          - Запустить бенчмарки"
	@echo "  profile            - Полный workflow профилирования"
	@echo "  collect-profiles   - Собрать профили (требует запущенного приложения)"
	@echo "  compare-profiles   - Сравнить профили base.pprof и result.pprof"
	@echo "  run-with-profiling - Запустить приложение с профилированием"
	@echo "  fmt                - Форматировать код (gofmt + goimports)"
	@echo "  fmt-check          - Проверить форматирование (для CI/CD)"
	@echo "  lint               - Запустить линтер (golangci-lint или go vet)"
	@echo "  check              - Комплексная проверка (форматирование + линтинг + тесты)"
	@echo "  clean              - Очистить профили и бинарники"
	@echo "  help               - Показать эту справку"