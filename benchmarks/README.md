# Benchmarks and Profiling

Система бенчмарков и профилирования для анализа производительности сервиса сокращения URL.

## Структура

- `profiles/` - Профили производительности и отчеты
- `utils/` - Утилиты для профилирования и генерации данных
- `component/` - Бенчмарки отдельных компонентов (будущее развитие)
- `integration/` - Интеграционные бенчмарки (будущее развитие)

## Быстрый старт

### Полный workflow (рекомендуется)
```bash
make profile
```

### Пошаговое профилирование
```bash
# 1. Собрать и запустить приложение
make build
make run-with-profiling

# 2. В другом терминале собрать профили
make collect-profiles

# 3. Применить оптимизации в коде

# 4. Собрать результирующие профили
curl http://localhost:6060/debug/pprof/heap > benchmarks/profiles/result.pprof

# 5. Сравнить профили
make compare-profiles
```

## Профилирование

### Включение профилирования в приложении

Профилирование можно включить через переменные окружения:
```bash
export PROFILING_ENABLED=true
export PROFILING_PORT=:6060
```

Или через флаги командной строки:
```bash
./shortener -profiling -profiling-port=:6060
```

### Сбор профилей

После запуска приложения с включенным профилированием:

```bash
# Профиль памяти
curl http://localhost:6060/debug/pprof/heap > profiles/base.pprof

# Профиль CPU (30 секунд)
curl http://localhost:6060/debug/pprof/profile?seconds=30 > profiles/cpu.pprof

# Профиль горутин
curl http://localhost:6060/debug/pprof/goroutine > profiles/goroutine.pprof

# Профиль блокировок
curl http://localhost:6060/debug/pprof/mutex > profiles/mutex.pprof
```

### Анализ профилей

```bash
# Анализ профиля памяти
go tool pprof profiles/base.pprof

# Сравнение профилей
go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof

# Веб-интерфейс для анализа
go tool pprof -http=:8081 profiles/base.pprof
```

## Автоматизированное профилирование

Для автоматического сбора и сравнения профилей используйте утилиты из `utils/`:

```go
import "github.com/Adigezalov/shortener/benchmarks/utils"

collector := utils.NewProfileCollector("profiles")

// Сбор базового профиля
err := collector.CollectHeapProfile("base.pprof")

// ... выполнение оптимизаций ...

// Сбор результирующего профиля
err = collector.CollectHeapProfile("result.pprof")
```

## Интерпретация результатов

### Бенчмарки

- `ns/op` - наносекунды на операцию (меньше = лучше)
- `B/op` - байт выделено на операцию (меньше = лучше)  
- `allocs/op` - количество аллокаций на операцию (меньше = лучше)

### Профили

При успешной оптимизации команда `pprof -diff_base` покажет:
- Отрицательные значения для оптимизированных функций
- Уменьшение общего потребления памяти
- Снижение количества аллокаций

## Best Practices

1. Запускайте бенчмарки на стабильной системе
2. Используйте `-benchtime` для более точных измерений
3. Собирайте профили под реальной нагрузкой
4. Сравнивайте профили до и после оптимизаций
5. Документируйте изменения и их влияние на производительность