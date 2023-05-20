# REST API Documentation

## Стек технологий

#### API
- Go
- Fiber
#### Хранилище:
- Postgres
#### Кеширование:
- Redis
#### Логирование:
- Zap

### Проблемы, возникшие при разработке, и их возможные решение:

1) **Проблема**: сайт ответил 429 Too Many Requests   
   **Решение**: данная проблема не была решена, как вариант, можно добавлять сайты, которые вернули такой ответ в *карантин* на n минут. Перед проверкой сайта на время ответа смотреть, не лежит ли этот сайт в карантине
2) **Проблема**: где и как хранить ссылки на сайты   
   **Решение**: было принято решение хранить ссылки на сайте в базе данных Postgres. Была создана таблица **website**, с полями **url** - ссылка на сайт без scheme, **last_check_at** - дата последней проверки доступности, **access_time** - время доступа к сайту, **available** - доступен ли сайт (сайт считается недоступным, если он вернул статус, отличный от 200). При необходимости можно добавить возможность добавлять ссылки через endpoints
3) **Проблема**: кеширование   
   **Решение**: все ответы на endpoints, которые могут иметь высокую нагрузку кешируются с помощью Redis  
4) **Проблема**: метрики   
   **Решение**: вместо использования Prometheus было принято решение написать свою простую оболочку для метрик на основе того же Redis

---

## Deployment

**Build** application

```shell
docker compose build
```

---

**Run** application

```shell
docker compose up -d
```

---

## Примеры использования

### Получить время доступа к определенному сайту
#### Запрос
```http request
HTTP/1.1 http://localhost:8080/api/v1/estimate?url=google.com
Accept: application/json  
```

#### Ответ
```json
{
  "last_check_at": "2023-05-20T14:53:54.320898+03:00",
  "access_time": "293.102ms"
}
```

---

### Получить имя сайта с минимальным временем доступа
#### Запрос
```http request
HTTP/1.1 GET http://localhost:8080/api/v1/estimate/min
Accept: application/json  
```

#### Ответ
```json
{
  "url": "login.tmall.com",
  "last_check_at": "2023-05-20T14:54:54.074799+03:00",
  "access_time": "48.651ms"
}
```

---

### Получить имя сайта с максимальным временем доступа
#### Запрос
```http request
HTTP/1.1 GET http://localhost:8080/api/v1/estimate/max
Accept: application/json  
```

#### Ответ
```json
{
  "url": "microsoft.com",
  "last_check_at": "2023-05-20T14:57:04.443525+03:00",
  "access_time": "10.417853s"
}
```

---

### Получить метрики по запросам
#### Запрос
```http request
HTTP/1.1 GET http://localhost:8080/admin/metrics  
Accept: application/json  
Authorization: Basic YWRtaW46YWRtaW4=  
```

#### Ответ
```json
[
  {
    "endpoint": "/api/v1/estimate/min",
    "count": 9
  },
  {
    "endpoint": "/api/v1/estimate/max",
    "count": 5
  },
  {
    "endpoint": "/api/v1/estimate",
    "count": 4
  }
]
```

---

## Конфигурации

### Все параметры загружаются из файта **[.env](.env)**

```dotenv
LOG_LEVEL=debug

SERVER_ADDR=:8080
SERVER_ADMIN_USERNAME=admin
SERVER_ADMIN_PASSWORD=admin

POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=estimate

REDIS_ADDR=redis:6379

ESTIMATION_PERIOD=1m
```