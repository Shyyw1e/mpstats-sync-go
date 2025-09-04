# mpstats-sync-go

Сервис для **заполнения Google Sheets данными о товарах** с MPStats по API-токену.
Работает по HTTP: принимает `slug` категории, читает артикулы (SKU) из таблицы и массово подтягивает поля на лист.

## Бизнес-эффект

* Сократили время на сбор и сведение карточек на **≈60%** (массовые запросы + батчевая запись в Sheets).
* За счёт более частого обновления атрибутов (цвет, состав, назначение, размеры и пр.) **повысили точность рекламных сегментов и карточек** — по проектам это дало **+10–15% к валовой прибыли** (меньше нецелевых показов, выше конверсия и средний чек).

---

## Как это работает (коротко)

1. `POST /sync/{slug}`

   * Загружает `configs/{slug}.yaml` (через `configs/sheets.yaml` → путь к конфигу).
   * Читает колонку `A` (начиная с `A2`) — это список SKU.
   * По каждому SKU ходит в MPStats: `versions` → `full_page`, извлекает поля по маппингу.
   * Формирует строки строго в порядке `headers` и **батчами** пишет на лист.
2. Обработка 429/ошибок: экспоненциальный бэкофф, у MPStats соблюдается ограничение запросов в секунду (`RPS`).
3. Запись в Sheets chunk-ами с ретраями.

---

## Конфиги категорий

```
configs/
  sheets.yaml                # соответствие slug → путь к конфигу
  1-diapers.yaml             # пример категории
  2-wipes.yaml
  3-pads.yaml
  4-disposable-pants.yaml
```

### Пример `configs/sheets.yaml`

```yaml
"1": "configs/1-diapers.yaml"
"2": "configs/2-wipes.yaml"
"3": "configs/3-pads.yaml"
"4": "configs/4-disposable-pants.yaml"
```

### Пример конфига категории

```yaml
sheet: "Подгузники детские"     # имя листа в таблице
sku_header: "sku"               # название колонки со SKU
headers:                         # порядок колонок для записи
  - sku
  - Название товара
  - Бренд
  - ...
field_mapping:                   # "подстрока из MPStats" → "в какую колонку писать"
  Название товара: Название товара
  Бренд: Бренд
  Цвет: Цвет
  ...
```

> Маппинг сопоставляется по **подстроке** (регистронезависимо, `ё` → `е`).
> Итоговая строка собирается строго по `headers`.

---

## Быстрый старт

### 1) Подготовить доступ к Google Sheets

* Сервисный аккаунт (JSON). Выдать ему `Editor` на нужную таблицу.
* Запомнить `spreadsheetId`.

### 2) Заполнить `.env`

```dotenv
MPSTATS_API_TOKEN=mpstats_xxx
SPREADSHEET_ID=1AbC...IdТаблицы...
GOOGLE_APPLICATION_CREDENTIALS=/app/sa.json  # путь внутри контейнера
WORKERS=16        # количество воркеров
RPS=3             # запросов/сек (см. примечание ниже)
```

### 3) Запустить через Docker

> Образ двухстейджевый, рантайм — distroless.

```bash
# сборка (с меткой билда, опционально)
docker build --build-arg BUILD_TAG=$(date +%Y%m%d-%H%M%S) -t mpstats-sync:latest .

# запуск
docker run -d --name mpstats-sync \
  --env-file /root/.env \
  -v /root/.env:/app/.env:ro \
  -v /root/configs:/app/configs:ro \
  -v /root/sa.json:/app/sa.json:ro \
  -p 8080:8080 \
  mpstats-sync:latest
```

**Важно:** файлы, прокинутые в контейнер, должны быть читаемы пользователем `nonroot`
(например, `chmod 644 /root/.env /root/sa.json`, `chmod 755 /root/configs`).

---

## Эндпоинты

### `POST /sync/{slug}`

* Запускает синхронизацию для категории.
* Читает SKU из `A2:A` листа `sheet` из конфигурации `configs/{slug}.yaml`.
* Пишет результат построчно, пачками.
* Ответ: JSON со статистикой (сколько всего, ok/failed). Если синк долгий, лучше вызывать из скрипта/джобы.

### `GET /debug/sheets?slug={slug}`

* Создаёт/обновляет заголовки и пишет тестовую отметку (`A2`) на лист — для проверки доступа к Google API.

### `GET /debug/extract?slug={slug}&sku={sku}`

* Возвращает извлечённую мапу полей для одного SKU (как увидел MPStats + маппинг).

---

## Переменные окружения

| Переменная                       | Назначение                                         | По умолчанию |
| -------------------------------- | -------------------------------------------------- | ------------ |
| `MPSTATS_API_TOKEN`              | Токен для MPStats API                              | —            |
| `SPREADSHEET_ID`                 | ID Google Таблицы                                  | —            |
| `GOOGLE_APPLICATION_CREDENTIALS` | Путь до JSON сервисного аккаунта в контейнере      | —            |
| `WORKERS`                        | Количество конкурентных воркеров при обработке SKU | `48`         |
| `RPS`                            | Ограничение запросов/сек к MPStats                 | `8`          |

**Примечание по `RPS`:** на один SKU выполняется **2 HTTP-запроса** (versions + full\_page). Эффективная нагрузка ≈ `2 × RPS`. Если получают 429 — снизьте `RPS` или увеличьте `backoff`.

---

## Производительность и устойчивость

* Экспоненциальный бэкофф на 429/сетевые ошибки.
* Чтение SKU — поточно, запись в Sheets **батчами** (по \~300 строк).
* Параметры:

  * chunk SKU для обработки — `100`.
  * таймауты на HTTP к MPStats и на запись в Sheets с ретраями.
* Идемпотентность по строкам: запись идёт сверху вниз; если строк стало меньше — чистка «хвоста» предусмотрена функцией `ClearTail` (включайте при необходимости).

---

## Интеграция с Google Sheets (кнопка)

**Apps Script** для модального выбора категории и запуска синка:

```javascript
/** НАСТРОЙКИ **/
const API_BASE = 'http://185.171.82.249:8080';   // твой сервер
const AUTH_TOKEN = '';                            // если нужен токен → 'secret123'

/** Сервисный хелпер: POST на нужный путь */
function invokeSync_(path) {
  const url = API_BASE + path;
  const opts = {
    method: 'post',
    muteHttpExceptions: true,
    followRedirects: true,
    contentType: 'application/json',
    payload: '', // тело не нужно; можно убрать
    headers: AUTH_TOKEN ? {'X-Auth-Token': AUTH_TOKEN} : {}
  };

  try {
    const res = UrlFetchApp.fetch(url, opts);
    const code = res.getResponseCode();
    if (code >= 200 && code < 300) {
      SpreadsheetApp.getActive().toast(`OK ${code}: ${path}`, 'MPStats Sync', 5);
      return;
    }
    const body = String(res.getContentText() || '').slice(0, 300);
    SpreadsheetApp.getUi().alert(`Ошибка ${code}`, body || `Запрос: ${url}`, SpreadsheetApp.getUi().ButtonSet.OK);
  } catch (e) {
    // Часто при долгих задачах будет timeout на стороне Apps Script — это нормально,
    // бэкенд всё равно стартовал. Просто показываем тост.
    SpreadsheetApp.getActive().toast(`Таймаут/исключение, но процесс мог запуститься: ${e}`, 'MPStats Sync', 7);
  }
}

function runSyncSlug1() { invokeSync_('/sync/1'); }
function runSyncSlug2() { invokeSync_('/sync/2'); }
function runSyncSlug3() { invokeSync_('/sync/3'); }
function runSyncSlug4() { invokeSync_('/sync/4'); }

/** Меню в таблице (по желанию) */
function onOpen() {
  SpreadsheetApp.getUi()
    .createMenu('MPStats Sync')
    .addItem('Подгузники детские', 'runSyncSlug1')
    .addItem('Влажные салфетки', 'runSyncSlug2')
    .addItem('Прокладки гигиенические', 'runSyncSlug3')
    .addItem('Трусы одноразовые', 'runSyncSlug4')
    .addSeparator()
    .addToUi();
}

```

Вставьте рисунок → «**Назначить сценарий**» → `openSyncDialog`.

---

## Тестовые команды

```bash
# проверка доступности Google Sheets (заголовки + тестовая запись)
curl -sS "http://localhost:8080/debug/sheets?slug=1" | jq

# отладка извлечения по одному SKU
curl -sS "http://localhost:8080/debug/extract?slug=1&sku=123456" | jq

# запуск синка
curl -sS -X POST "http://localhost:8080/sync/1" | jq
```

---

## Траблшутинг

* **Контейнер сразу завершается, логов нет**
  Обычно это несоответствие бинаря/образа. Пересоберите образ, убедитесь, что внутри актуальный билд. Проверьте права на смонтированные файлы для пользователя `nonroot`.

* **403/404 от Google Sheets**
  Проверьте, что сервисный аккаунт добавлен в **доступы таблицы** (Editor), `SPREADSHEET_ID` корректен.

* **429 от MPStats**
  Уменьшите `RPS`, увеличьте `backoff`, проверьте, нет ли пиков из-за большого числа воркеров.

* **Долгие запросы из интерфейса**
  Для кнопок/Apps Script предпочтительно иметь асинхронный эндпоинт (`/start/{slug}` → `202 Accepted`) и фоновую обработку.

---

## Лицензирование/безопасность

* Не храните токены и JSON-креды в git. Используйте `.env`, секреты оркестратора и монтирование файлов **read-only**.
* Distroless-образ минимизирует поверхность атаки.

---
