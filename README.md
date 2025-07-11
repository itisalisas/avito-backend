# avito-backend

## Инструкция по запуску

Необходим установленный Go, PostgreSQL и Docker.

Скачайте репозиторий:

```shell
git clone https://github.com/itisalisas/avito-backend.git
cd avito-backend
```

Укажите в `.env` файле необходимые переменные окружения 
(если тесты на БД не нужны, переменные с суффиксом `_TEST` можно не задавать)

Запустите сервис:

```shell
docker-compose up --build pvz-service
```

Для запуска тестов:

```shell
docker-compose up --build tests
```

В проекте настроена кодогенерация DTO endpoint'ов по openapi схеме
Для запуска:
```shell
make generate-dto
````

## Выполненные задания

Помимо основных заданий, из дополнительных заданий выполнено:
- реализована пользовательская авторизация по методам /register и /login 
- добавлен prometheus и сбор метрик
- настроена кодогенерация DTO endpoint'ов по openapi схеме
- из логирования логируются коды коврата через `middleware.Logger`, в теории можно было еще добавить `zap`
- реализован gRPC-метод, который возвращает все добавленные в систему ПВЗ. Можно попробовать, запустив сервер и запустив
```shell
grpcurl -plaintext localhost:3000 pvz.v1.PVZService/GetPVZList
```

Немного не хватило времени, хотелось настроить нормальный запуск тестов, с настройкой запуска тестов на БД 
через .env не успела справиться, поэтому они там падают, про in-memory БД типо H2 для Java не нашла ничего(. 
Можно запустить локально, указав нужные поля.