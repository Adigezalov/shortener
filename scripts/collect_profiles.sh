#!/bin/bash

# Скрипт для автоматического сбора профилей производительности

set -e

# Конфигурация
PROFILING_URL="http://localhost:6060"
PROFILES_DIR="benchmarks/profiles"
PROFILE_DURATION=30

# Создаем директорию для профилей если она не существует
mkdir -p "$PROFILES_DIR"

echo "Collecting performance profiles..."

# Проверяем доступность сервера профилирования
if ! curl -s "$PROFILING_URL/debug/pprof/" > /dev/null; then
    echo "Error: Profiling server is not available at $PROFILING_URL"
    echo "Make sure the application is running with profiling enabled:"
    echo "  export PROFILING_ENABLED=true"
    echo "  ./shortener -profiling"
    exit 1
fi

echo "Profiling server is available"

# Собираем базовый профиль памяти
echo "Collecting heap profile..."
curl -s "$PROFILING_URL/debug/pprof/heap" > "$PROFILES_DIR/base.pprof"
echo "Heap profile saved to $PROFILES_DIR/base.pprof"

# Собираем профиль CPU
echo "Collecting CPU profile (${PROFILE_DURATION}s)..."
curl -s "$PROFILING_URL/debug/pprof/profile?seconds=$PROFILE_DURATION" > "$PROFILES_DIR/cpu.pprof"
echo "CPU profile saved to $PROFILES_DIR/cpu.pprof"

# Собираем профиль горутин
echo "Collecting goroutine profile..."
curl -s "$PROFILING_URL/debug/pprof/goroutine" > "$PROFILES_DIR/goroutine.pprof"
echo "Goroutine profile saved to $PROFILES_DIR/goroutine.pprof"

# Собираем профиль блокировок
echo "Collecting mutex profile..."
curl -s "$PROFILING_URL/debug/pprof/mutex" > "$PROFILES_DIR/mutex.pprof"
echo "Mutex profile saved to $PROFILES_DIR/mutex.pprof"

# Собираем профиль аллокаций
echo "Collecting allocs profile..."
curl -s "$PROFILING_URL/debug/pprof/allocs" > "$PROFILES_DIR/allocs.pprof"
echo "Allocs profile saved to $PROFILES_DIR/allocs.pprof"

echo "Profile collection completed!"
echo ""
echo "To analyze profiles:"
echo "  go tool pprof $PROFILES_DIR/base.pprof"
echo "  go tool pprof -http=:8081 $PROFILES_DIR/base.pprof"
echo ""
echo "To compare profiles after optimization:"
echo "  go tool pprof -top -diff_base=$PROFILES_DIR/base.pprof $PROFILES_DIR/result.pprof"