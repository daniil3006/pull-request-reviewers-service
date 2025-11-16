**Запуск приложения**  

**Клонирование репозитория**
```bash
git clone https://github.com/daniil3006/pull-request-reviewers-service.git
cd pull-request-reviewers-service
```

**Запуск в Docker**  
В корне проекта выполнить:
```bash
docker compose up --build
```

Сервис доступен по адрессу:
```bash
http://localhost:8080
```

PostgreSQL доступен по адрессу:
```bash
localhost:5433
```

Добавлен эндпоинт, возвращающий количество назначений по пользователям  
**GET** /stats/reviewers  
**Response**
 -`200 OK`
