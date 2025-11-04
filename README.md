# Order Manager

Сервис для управления заказами.

## Настройка и запуск

### Предварительные требования

- [Go](https://golang.org/)
- [Make](https://www.gnu.org/software/make/)
- [Protoc](https://grpc.io/docs/protoc-installation/)

### Установка

1. Клонируйте репозиторий:
   ```bash
   git clone <repository-url>
   ```
2. Перейдите в директорию проекта:
   ```bash
   cd order-manager
   ```
3. Установите зависимости:
   ```bash
   go mod tidy
   ```

### Конфигурация

Приложение можно настроить с помощью переменных окружения или файла конфигурации.

#### Переменные окружения

- `ENV`: Среда выполнения (например, `dev`, `prod`). **Обязательная**.
- `GRPC_PORT`: Порт для gRPC сервера (по умолчанию `8081`).
- `GATEWAY_PORT`: Порт для gRPC шлюза (по умолчанию `8080`).
- `GRPC_TIMEOUT`: Таймаут для gRPC соединений (по умолчанию `5s`).
- `CONFIG_PATH`: Путь к файлу конфигурации `settings.yml`.

#### Файл конфигурации

Вы можете использовать файл `config/settings.yml` для задания конфигурации.

Пример `config/settings.yml`:
```yaml
env: "dev"

grpc:
  port: 50051
  timeout: 1h

gateway:
   port: 8080
```

### Запуск приложения

1. **Сборка и запуск:**
   ```bash
   make run
   ```
   Эта команда соберет и запустит приложение. По умолчанию приложение будет использовать переменные окружения.

2. **Запуск с файлом конфигурации:**
   Вы можете указать путь к файлу конфигурации при запуске:
   ```bash
   ./bin/order-manager --config=./config/settings.yml
   ```
   Или через `Makefile`:
   ```bash
   make run-config
   ```

### Пример запуска с переменными окружения

```bash
export ENV="dev"
export GRPC_PORT=50051
export GATEWAY_PORT=8080
export GRPC_TIMEOUT=10s
make run
```
