# Уведомления о Дне Рождения

Данный сервис отправляет уведомления пользователям о днях рождения из их подписок.

## Использование API

### Регистрация пользователя

**URL:** `/api/registration`  
**Метод:** `POST`  
**Описание:** Регистрирует нового пользователя.

**Пример запроса:**
```sh
curl -X POST http://localhost:8080/api/registration \
    -H "Content-Type: application/json" \
    -d '{
          "name": "John Doe",
          "email": "johndoe@example.com",
          "password": "password123",
          "date_of_birth": "1990-01-01"
        }'
```

### Вход в систему

**URL:** `/api/login`  
**Метод:** `POST`  
**Описание:** Авторизует пользователя и возвращает JWT токен.


**Пример запроса:**
```sh
curl -X POST http://localhost:8080/api/login \
    -H "Content-Type: application/json" \
    -d '{
          "email": "johndoe@example.com",
          "password": "password123"
        }'

```




### Подписка на пользователя


**URL:** `/api/subscribe`  
**Метод:** `POST`  
**Описание:** Создает подписку на другого пользователя. Требуется JWT токен в заголовке Authorization


**Пример запроса:**
```sh
curl -X POST http://localhost:8080/api/subscribe \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <JWT_TOKEN>" \
    -d '{
          "related_user_id": 2
        }'

```

### Отписка от пользователя


**URL:** `/api/unsubscribe`  
**Метод:** `POST`  
**Описание:** Удаляет подписку на пользователя. Требуется JWT токен в заголовке Authorization


**Пример запроса:**
```sh
curl -X POST http://localhost:8080/api/unsubscribe \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <JWT_TOKEN>" \
    -d '{
          "related_user_id": 2
        }'

```

### Получение доступных для подписки пользователей


**URL:** `/api/available`  
**Метод:** `GET`  
**Описание:** Возвращает список пользователей, на которых текущий пользователь еще не подписан. Требуется JWT токен в заголовке Authorization


**Пример запроса:**
```sh
curl -X GET http://localhost:8080/api/available \
-H "Authorization: Bearer <JWT_TOKEN>"
```
