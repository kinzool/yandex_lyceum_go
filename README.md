# Финальная задача
## Описание
Я реализовал веб-сервер на языке Go, который принимает POST-запрос с параметром "expression" в теле запроса, которое необходимо будет рассчитать с помощью ранее написанной мной программы калькулятор.
## Инструкция по запуску
Для того, чтобы запустить сервер, необходимо в консоль ввести:
```
go run .\cmd\main.go
```
Отправить POST-запрос серверу можно отправить с помощью curl:
```
curl --location 'localhost:8080/api/v1/calculate' \         
  --header 'Content-Type: application/json' \         
  --data '
  {
    "expression": "Введите своё арифметическое выражение"
  }'
```
## Примеры запросов
### Успешный запрос
```
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2*2"
}'
```
### Запрос, возвращающий ошибку с кодом 422
```
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2/-"
}'
```
### Запрос, возвращающий ошибку с кодом 500
```
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2/0"
}'
```
