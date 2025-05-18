# Список команд для теста:

Файл для импорта
data/import_demo.json

```json
[
  { "order_id": "ORD200", "user_id": "U999", "expires_at": "2025-07-15" },
  { "order_id": "ORD201", "user_id": "U999", "expires_at": "2025-07-15" }
]
```

| №      | Команда                                                                                                                                                   | Ожидаемый вывод / комментарий                                                                                                                                 |
|--------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **0**  | `help`                                                                                                                                                    | Появится сводка всех команд.                                                                                                                                  |
| **1**  | `accept-order --order-id ORD001 --user-id U123 --expires 2025-06-30`                                                                                      | `ORDER_ACCEPTED: ORD001`                                                                                                                                      |
| **2**  | *(дубликат)*<br>`accept-order --order-id ORD001 --user-id U123 --expires 2025-06-30`                                                                      | `ERROR: ORDER_ALREADY_EXISTS: order already exists`                                                                                                           |
| **3**  | *(прошедшая дата)*<br>`accept-order --order-id ORD002 --user-id U123 --expires 2020-01-01`                                                                | `ERROR: VALIDATION_FAILED: validation failed`                                                                                                                 |
| **4**  | `accept-order --order-id ORD003 --user-id U123 --expires 2025-06-30`                                                                                      | `ORDER_ACCEPTED: ORD003`                                                                                                                                      |
| **5**  | `list-orders --user-id U123 --in-pvz`                                                                                                                     | две строки `ORDER: …` + `TOTAL: 2`                                                                                                                            |
| **6**  | `process-orders --user-id U123 --action issue --order-ids ORD001,ORD003`                                                                                  | `PROCESSED: ORD001`<br>`PROCESSED: ORD003`                                                                                                                    |
| **7**  | `process-orders --user-id U123 --action issue --order-ids ORD001`                                                                                         | `ERROR ORD001: VALIDATION_FAILED` (уже выданы)                                                                                                                |
| **8**  | `accept-order --order-id ORD004 --user-id U124 --expires 2025-06-30`                                                                                      | `ORDER_ACCEPTED: ORD004`                                                                                                                                      |
| **9**  | *(клиентский возврат < 48 ч)*<br>`process-orders --user-id U123 --action return --order-ids ORD001`                                                       | `PROCESSED: ORD001`                                                                                                                                           |
| **10** | `list-returns`                                                                                                                                            | `RETURN: ORD001 U123 <date>` + `PAGE: 1 LIMIT: 20`                                                                                                            |
| **11** | Завершите приложение, измените  дату руками в файле orders.json `expires_at` на вчера для заказа с id=ORD004, затем:<br/>`return-order --order-id ORD004` | `ORDER_RETURNED: ORD004`                                                                                                                                      |
| **12** | `list-returns --page 1 --limit 1`                                                                                                                         | постранично один возврат                                                                                                                                      |
| **13** | `order-history`                                                                                                                                           | лента `HISTORY: …` (ACCEPTED → ISSUED → RETURNED …)                                                                                                           |
| **14** | `import-orders --file data/import_demo.json`                                                                                                              | `IMPORTED: 2`                                                                                                                                                 |
| **15** | `list-orders --user-id U999 --page 1 --limit 5`                                                                                                           | увидите ORD200, ORD201                                                                                                                                        |
| **16** | `scroll-orders --user-id U999 --limit 1`                                                                                                                  | CLI покажет первую запись + `NEXT: ORD200` <br>введите `next` → вторая запись + `NEXT: ORD201` <br>ещё `next` → приложение завершит цикл (больше нет данных). |


