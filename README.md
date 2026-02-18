# WeRide — backend для совместных поездок в такси

## Архитектура

```
                    ┌──────────────┐
  HTTP :8080  ────► │  API Gateway │
                    └──────┬───────┘
                           │ gRPC
          ┌────────────────┼──────────────────┐
          ▼                ▼                  ▼
  ┌──────────────┐ ┌──────────────┐ ┌──────────────────┐
  │ User Service │ │ Room Service │ │ Payment Service  │
  │   :50052     │ │   :50051     │ │     :50053       │
  └──────┬───────┘ └──────┬───────┘ └────────┬─────────┘
         └────────────────┴──────────────────┘
                           │ SQL
                    ┌──────▼───────┐
                    │  PostgreSQL  │
                    │  3 databases │
                    │  :5432       │
                    └──────────────┘
```

## Быстрый старт

```bash
cd docker
cp .env.example .env
# Заполни .env: JWT_SECRET, YOOKASSA_SHOP_ID, YOOKASSA_SECRET_KEY
docker-compose up --build
```

API доступен на `http://localhost:8080`

---

## API Reference

### Auth
| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/auth/register` | Регистрация |
| POST | `/auth/login` | Вход, возвращает JWT |
| GET  | `/auth/history` | 🔒 История поездок |

### Rooms
| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/rooms` | 🔒 Создать комнату |
| GET  | `/rooms` | 🔒 Найти доступные |
| GET  | `/rooms/:id` | 🔒 Детали комнаты |
| POST | `/rooms/:id/join` | 🔒 Вступить |
| POST | `/rooms/:id/exit` | 🔒 Покинуть |
| POST | `/rooms/:id/complete` | 🔒 Завершить поездку (триггерит оплату) |

### Payments
| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/payments/process` | 🔒 Списать оплату |
| POST | `/payments/refund` | 🔒 Возврат при отмене |
| GET  | `/payments/history` | 🔒 История транзакций |

🔒 — требует заголовок `Authorization: Bearer <JWT>`

---

## Флоу поездки

```
1. POST /auth/register + POST /auth/login  → получаем JWT
2. POST /rooms                             → водитель создаёт комнату
3. POST /rooms/:id/join                    → пассажиры вступают
4. POST /rooms/:id/complete                → водитель завершает поездку
   └── автоматически:
       ├── сохраняет маршрут в user_service
       └── списывает cost_per_member с каждого пассажира через ЮKassa
5. GET  /payments/history                  → пассажир видит транзакцию
6. GET  /auth/history                      → история поездок
```

---

## ЮKassa

Получить credentials: https://yookassa.ru/my/api-keys

В `.env`:
```
YOOKASSA_SHOP_ID=ваш_shop_id
YOOKASSA_SECRET_KEY=ваш_secret_key
```

---

## CI/CD

Пайплайн в `.gitlab-ci.yml`:

```
lint → test → build → push → deploy (manual)
```

Переменные в GitLab (Settings → CI/CD → Variables):
- `JWT_SECRET`
- `YOOKASSA_SHOP_ID`
- `YOOKASSA_SECRET_KEY`
- `DEPLOY_HOST` — IP сервера
- `DEPLOY_USER` — SSH пользователь
- `DEPLOY_SSH_KEY` — приватный SSH ключ (masked)
