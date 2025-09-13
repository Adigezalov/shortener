.PHONY: build test benchmark profile clean help

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
	@echo "  clean              - Очистить профили и бинарники"
	@echo "  help               - Показать эту справку"