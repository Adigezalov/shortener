#!/bin/bash

# Скрипт для сравнения профилей до и после оптимизации

set -e

PROFILES_DIR="benchmarks/profiles"
BASE_PROFILE="$PROFILES_DIR/base.pprof"
RESULT_PROFILE="$PROFILES_DIR/result.pprof"

# Проверяем наличие базового профиля
if [ ! -f "$BASE_PROFILE" ]; then
    echo "Error: Base profile not found at $BASE_PROFILE"
    echo "Run ./scripts/collect_profiles.sh first to collect base profile"
    exit 1
fi

# Проверяем наличие результирующего профиля
if [ ! -f "$RESULT_PROFILE" ]; then
    echo "Error: Result profile not found at $RESULT_PROFILE"
    echo "Collect result profile after applying optimizations:"
    echo "  curl http://localhost:6060/debug/pprof/heap > $RESULT_PROFILE"
    exit 1
fi

echo "Comparing profiles..."
echo "Base profile: $BASE_PROFILE"
echo "Result profile: $RESULT_PROFILE"
echo ""

# Сравниваем профили
echo "=== Memory Usage Comparison ==="
go tool pprof -top -diff_base="$BASE_PROFILE" "$RESULT_PROFILE"

echo ""
echo "=== Detailed Comparison ==="
go tool pprof -list=. -diff_base="$BASE_PROFILE" "$RESULT_PROFILE"

echo ""
echo "Comparison completed!"
echo ""
echo "Negative values indicate memory usage reduction (optimization success)"
echo "Positive values indicate memory usage increase"