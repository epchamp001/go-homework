# Quickstart & Test Guide

## Запуск приложения

```bash
make start
```

После запуска откройте Swagger UI:

```bash
http://localhost:8080/swagger/http/index.html
```

## Порядок команд 

1. Приём одного заказа /v1/orders/accept
2. Импорт нескольких заказов /v1/orders/import
3. Просмотр списка заказов /v1/orders, ставим user_id=200
4. Откройте pgAdmin (http://localhost:5050), войдите:
   * Email: 123ozon123
   * Password: mega_secret_pass_ozon_dev
    Выберите сервер Postgres Master → база pvz → Query Tool.
    Выполните:
    ```sql
    UPDATE orders
    SET expires_at = now() - interval '1 day'
    WHERE id = 100;
    ```

5. Возврат заказа /v1/orders/100/return
6. Проверка возвратов /v1/orders/returns
7. Проверка истории /v1/orders/history
8. Ручка /v1/orders/process с телом:
    ```json
    {
    "user_id": 201,
    "action": "ACTION_TYPE_ISSUE",
    "order_ids": [101]
    }
    ```

9. Генерация отчёта клиентов : в браузере переходим по ссылке:
    ```bash
    http://localhost:8080/v1/reports/clients?sortBy=orders-
    ```