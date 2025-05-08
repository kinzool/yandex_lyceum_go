# Распределённый вычислитель арифметических выражений
## Описание
Я реализовал веб-сервер на языке Go, который принимает POST- и GET- запросы в endpoint'ах "/login", "/register", "/calculate", "/expressions", 
"/expressions/id", каждый из которых выполняет определенный функционал, соответствующий условиям задачи. Эта программа позволяет персистентно и многопользовательски распределенно вычислять арифметические выражения.
В этой версии приложения используется база данных sqlite, в которой создаются две таблицы: users и expressions. 

В таблице users хранятся данные о зарегистрированных пользователях в столбцах с названиями id, login и password. Пароль хранится в хешированном виде, что позволяет сохранять безопасность.

В таблице expressions хранятся выражения, которые добавляюся пользователями в столбцах с названиями id, user_id, expression, result.

-------------------------------------------------------------------------------------------------------
## Инструкция по запуску
## Для того, чтобы запустить сервер, необходимо:
### 1) Склонировать репозиторий
```
git clone https://github.com/kinzool/yandex_lyceum_go
cd yandex_lyceum_go
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

## Архитектура приложения (как все работает)
**Оркестратор** (порт 8080 по умолчанию):

- Принимает выражения через REST API
- Разбивает выражения на атомарные задачи
- Управляет очередью задач
- Собирает результаты
- Хранит статусы вычислений

**Агент**:

- Получают задачи через HTTP-запросы
- Выполняют арифметические операции с задержкой
- Возвращают результаты через API

### Агент

- `ORCHESTRATOR_URL` - URL оркестратора
- `COMPUTING_POWER` - количество параллельных задач

# Также можно запустить программу с помощью Docker. Для этого необходимо ввести следующую команду:
```
docker-compose up --build
```
## Server Endpoints
### 1) Регистрация пользователя (POST /api/v1/register)
### Пример запроса: 
```curl --location 'localhost:8080/api/v1/register' \
--header 'Content-Type: application/json' \
--data '{
	"login": "1",
    "password": "1"
}'
```
### Получаем ответ с кодом 200:
```
succesfull registration
```
Если метод запроса будет неправильным, то получим ошибку с кодом 405:
```
{"error":"Wrong Method"}
```

### 2) Вход пользователя (POST /api/v1/login)
### Пример запроса: 
```
curl --location 'localhost:8080/api/v1/login' \
--header 'Content-Type: application/json' \
--data '{
    "login": "1",
    "password": "1"
}'
```

### Получаем ответ с кодом 200 и JWT, который в дальнейшем будет использоваться для аутентификации пользователя с помощью AuthMiddleware(JWT хранится в Cookie):
```
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDY3MDI4NzYsImlhdCI6MTc0NjcwMjI3NiwibG9naW4iOiIxIiwibmJmIjoxNzQ2NzAyMjgxLCJ1c2VyX2lkIjoxfQ.sI4G6BJPLBpRFhywQ2_hYnRU69mssKoNL99nof7sBDQ"
}
```

### В случае неправильно введенных данных, получим ошибку с кодом 401 и ответ:
```
{"error":"Invalid credentials"}

```
## Все последующие запросы выполняются в контексте авторизованного пользователя. То есть, при получении всех выражений или выражения по идентификатору в ответе будут только те выражения, которые добавлял авторизованный пользователь.


### 3) Добавление выражения (POST /api/v1/calculate)
### Пример запроса:
```
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--header 'Cookie: auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDY3MzkyODYsImlhdCI6MTc0NjczODY4NiwibG9naW4iOiIyIiwibmJmIjoxNzQ2NzM4NjkxLCJ1c2VyX2lkIjoyfQ.h2uOIrAMBZ3RmNZuwCQiy3FmPNPKqTEaC3ouH7_O450' \
--data '{
    "expression": "2+1"
}'
```
### Получаем ответ с кодом 201:
```
{
    "id": "1"
}
```
Если метод запроса будет неправильным, то получим ошибку с кодом 405:
```
{"error":"Wrong Method"}
```
Если пользователь не авторизован, получим ошибку с кодом 401 и ответ:
```
{"error":"Missing token"}
```
------------------------------------------------------------------------------------
## Этот ответ будет универсален для всех запросов от не авторизованных пользователей
------------------------------------------------------------------------------------

### 4) Получение списка выражений (GET /api/v1/expressions)

#### Пример запроса:
```
curl --location 'localhost:8080/api/v1/expressions' \
--header 'Cookie: auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDY3MzkyODYsImlhdCI6MTc0NjczODY4NiwibG9naW4iOiIyIiwibmJmIjoxNzQ2NzM4NjkxLCJ1c2VyX2lkIjoyfQ.h2uOIrAMBZ3RmNZuwCQiy3FmPNPKqTEaC3ouH7_O450'
```

### Получаем ответ с кодом 200:
```
{
    "expressions": [
        {
            "id": "1",
            "result": 3
        }
    ]
}
```
Если метод запроса будет неправильным, то получим ошибку с кодом 405:
```
{"error":"Wrong Method"}
```

### 3) Получение выражения по ID (GET /api/v1/expressions/{id})
### Пример запроса:
```
curl --location 'localhost:8080/api/v1/expressions/1' \
--header 'Cookie: auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDY3MzkyODYsImlhdCI6MTc0NjczODY4NiwibG9naW4iOiIyIiwibmJmIjoxNzQ2NzM4NjkxLCJ1c2VyX2lkIjoyfQ.h2uOIrAMBZ3RmNZuwCQiy3FmPNPKqTEaC3ouH7_O450'
```
Если будет введен несуществующий id, то получим ошибку с кодом 404:
```
{"error":"Expression not found"}
```
Если же введенный ID существует, то получаем:
```
{
    "expression": {
        "id": "1",
        "result": 3
    }
}
```
## Agent
### 1. Получение задачи
```
GET /internal/task
```
Пример ответа с кодом 200:
```
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
  "result": 3
}
```
## Тестирование
Моя программа покрыта модульными и интеграционными тестами, для запуска которых необходимо в консоль прописать команды:
### Модульные
```
go test -v .\tests\module
```
### Интеграционные
```
go test -v .\tests\integration
```
