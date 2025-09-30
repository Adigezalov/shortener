# Информация о сборке

При старте приложение выводит информацию о сборке в следующем формате:

```
Build version: <версия>
Build date: <дата сборки>
Build commit: <хеш коммита>
```

## Установка значений при сборке

### Обычная сборка
```bash
make build
```
Вывод:
```
Build version: N/A
Build date: N/A
Build commit: N/A
```

### Сборка с информацией о версии
```bash
VERSION=1.0.0 make build-with-version
```
Вывод:
```
Build version: 1.0.0
Build date: 2025/09/28 20:12:39
Build commit: 0f95a9f
```

### Ручная сборка с ldflags
```bash
go build -ldflags="-X 'main.buildVersion=1.2.3' -X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')' -X 'main.buildCommit=$(git rev-parse --short HEAD)'" -o shortener ./cmd/shortener
```

## Переменные сборки

В пакете `cmd/shortener` определены следующие глобальные переменные:

- `buildVersion` - версия сборки (по умолчанию "N/A")
- `buildDate` - дата и время сборки (по умолчанию "N/A") 
- `buildCommit` - короткий хеш Git коммита (по умолчанию "N/A")

Эти переменные устанавливаются во время сборки с помощью флага `-ldflags` и опции `-X`.

## Примеры значений

### Релизная сборка
```
Build version: v1.2.3
Build date: 2025/09/28 15:30:45
Build commit: a1b2c3d
```

### Сборка без Git
```
Build version: dev
Build date: 2025/09/28 15:30:45
Build commit: unknown
```

### Сборка разработчика
```
Build version: N/A
Build date: N/A
Build commit: N/A
```
