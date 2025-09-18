#!/bin/bash

# Полный workflow для бенчмарков и профилирования

set -e

PROFILES_DIR="benchmarks/profiles"
APP_BINARY="./shortener"
PROFILING_PORT=":6060"

echo "=== Benchmark and Profiling Workflow ==="
echo ""

# Проверяем, что приложение собрано
if [ ! -f "$APP_BINARY" ]; then
    echo "Building application..."
    go build -o shortener ./cmd/shortener
fi

# Создаем директорию для профилей
mkdir -p "$PROFILES_DIR"

echo "Step 1: Starting application with profiling enabled..."
export PROFILING_ENABLED=true
export PROFILING_PORT="$PROFILING_PORT"

# Запускаем приложение в фоне
$APP_BINARY &
APP_PID=$!

# Функция для остановки приложения при выходе
cleanup() {
    echo "Stopping application..."
    kill $APP_PID 2>/dev/null || true
    wait $APP_PID 2>/dev/null || true
}
trap cleanup EXIT

# Ждем запуска приложения
echo "Waiting for application to start..."
sleep 3

# Проверяем доступность профилирования
PROFILING_URL="http://localhost${PROFILING_PORT#:}"
if ! curl -s "$PROFILING_URL/debug/pprof/" > /dev/null; then
    echo "Error: Profiling server is not available"
    exit 1
fi

echo "Step 2: Collecting base profile..."
./scripts/collect_profiles.sh

echo ""
echo "Step 3: Generating load for profiling..."
# Генерируем нагрузку на приложение
for i in {1..500}; do
    curl -s -X POST -H "Content-Type: text/plain" -H "Accept-Encoding: gzip" \
         -d "https://example.com/test$i" \
         http://localhost:8080/ > /dev/null || true
done

echo "Load generation completed"

echo ""
echo "Step 4: Collecting result profile (after load)..."
curl -s "$PROFILING_URL/debug/pprof/heap" > "$PROFILES_DIR/result.pprof"
echo "Result profile saved to $PROFILES_DIR/result.pprof"

echo ""
echo "Step 5: Comparing profiles..."
if [ -f "$PROFILES_DIR/base.pprof" ] && [ -f "$PROFILES_DIR/result.pprof" ]; then
    ./scripts/compare_profiles.sh
else
    echo "Profiles not found for comparison"
fi

echo ""
echo "=== Workflow completed! ==="
echo ""
echo "Next steps:"
echo "1. Analyze profiles to identify optimization opportunities"
echo "2. Apply optimizations to the code"
echo "3. Run this workflow again to measure improvements"
echo ""
echo "Profile analysis commands:"
echo "  go tool pprof $PROFILES_DIR/base.pprof"
echo "  go tool pprof -http=:8081 $PROFILES_DIR/base.pprof"