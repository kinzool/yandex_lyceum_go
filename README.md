# Распределённый вычислитель арифметических выражений
## Описание
Я реализовал веб-сервер на языке Go, который принимает POST- и GET- запросы в endpoint'ах "/calculate", "/expressions", 
"/expressions/id", каждый из которых выполняет определенный функционал, соответствующий условиям задачи. Эта программа позволяет распределенно вычислять арифметические выражения
## Инструкция по запуску
## Для того, чтобы запустить сервер, необходимо:
### 1) Склонировать репозиторий
```
git clone https://github.com/kinzool/yandex_lyceum_go
cd calc_go
```
### 2) Запустить оркестратор и установить значения для переменных среды
```
# Установка времени различных операций (в миллисекундах)
$env:TIME_ADDITION_MS = "200"
$env:TIME_SUBTRACTION_MS = "200"
$env:TIME_MULTIPLICATIONS_MS = "300"
$env:TIME_DIVISIONS_MS = "400"

# Запуск оркестратора
go run .\cmd\orchestrator\main.go
```
### 3) Запустить агент
```
# Указание вычислительной мощности (количество горутин) и URL оркестратора
$env:COMPUTING_POWER = "4"
$env:ORCHESTRATOR_URL = "http://localhost:8080"

# Запуск агента
go run .\cmd\agent\main.go
```
## Переменные окружения
### Оркестратор

- `PORT` - порт сервера (по умолчанию 8080)
- `TIME_ADDITION_MS` - время сложения (мс)
- `TIME_SUBTRACTION_MS` - время вычитания (мс)
- `TIME_MULTIPLICATIONS_MS` - время умножения (мс)
- `TIME_DIVISIONS_MS` - время деления (мс)

### Агент

- `ORCHESTRATOR_URL` - URL оркестратора
- `COMPUTING_POWER` - количество параллельных задач

## Server Endpoints
### 1) Добавление выражения (POST /api/v1/calculate)
#### Пример запроса:
```
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "(2+2)/2*2"
}'
```
### Получаем ответ с кодом 201:
```
{
    "id": "1"
}
```

### 2) Получение списка выражений (GET /api/v1/expressions)
### Получаем ответ с кодом 200:
```
{
    "expressions": [
        {
            "id": "1",
            "status": "completed",
            "result": 4
        }
    ]
}
```
### 3) Получение выражения по ID (GET /api/v1/expressions/{id})
Если будет введен несуществующий id, то получим:
```
{"error":"Expression not found"}
```
Если же введенный ID существует, то получаем:
```
{
    "expression": {
        "id": "1",
        "status": "completed",
        "result": 4
    }
}
```
## Agent
### 1. Получение задачи
```
GET /internal/task
```
Пример ответа (200):
```json
{
    "task": {
        "id": "1",
        "arg1": 2,
        "arg2": 2,
        "operation": "+",
        "operation_time": 200
    }
}
```
### 2. Отправка результата
```
POST /internal/task
```
Пример запроса:

```
{
  "id": "1",
  "result": 4
}
```
## Тестирование
Моя программа покрыта тестами, для запуска которых необходимо в консоль прописать команду:
### Для оркестратора
```
go test .\tests\application\orchestrator_test.go
```
### Для агента
```
go test .\tests\calculator\calculator_test.go
```