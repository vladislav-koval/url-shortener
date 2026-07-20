# CLAUDE.md

Стабильные правила и знания о проекте `url-shortener`. Верифицировано чтением всего репозитория на момент написания — если что-то здесь противоречит коду, доверяй коду и поправь этот файл.

## Что это за проект

Укорачиватель ссылок на Go: создание короткого кода для URL, редирект по короткому коду, кеширование в Redis, публикация событий о переходах в Kafka и их обработка отдельным consumer'ом в фиче `analytics`, которая сохраняет клики в Postgres для последующей аналитики (сама аналитика — просмотр статистики — ещё не реализована).

Учебный/pet-проект одного разработчика, не production-нагрузка. Решения по масштабированию (Kafka `Async: true` с неограниченной внутренней очередью, отсутствие read-реплик и т.д.) осознанно упрощены под этот масштаб — см. обоснования в `docs/AI_HANDOFF.md`, не считай их автоматически багами, которые нужно чинить без запроса.

## Стек

- Go 1.26.1 (см. `go.mod`)
- Postgres 18 (`jackc/pgx/v5`) — источник истины для ссылок (`urlshortener.links`) и кликов (`urlshortener.clicks`)
- Redis 8 (`redis/go-redis/v9`) — read-through/write-through кеш поверх Postgres для ссылок
- Kafka 4.3, KRaft, один брокер (`segmentio/kafka-go`) — публикация и обработка событий кликов
- `go.uber.org/zap` — логирование
- `kelseyhightower/envconfig` — конфигурация из переменных окружения
- `go-playground/validator/v10` — валидация DTO по тегам
- `golang-migrate` (через Docker-образ `migrate/migrate`) — миграции Postgres
- Тестов нет вообще (`find . -name "*_test.go"` — пусто). Линтера кроме `go vet`/`gofmt` нет.

## Архитектурные правила

### `internal/core/` vs `internal/features/`

- `core/` — технология-специфичная, не знающая о бизнес-логике инфраструктура: логгер, HTTP-транспорт (`core/transport/http/{middleware,request,response,server}`), доступ к Postgres/Redis/Kafka, домен (`core/domain`), общие ошибки (`core/errors`), общий событийный контракт (`core/messaging/events`).
- `features/<feature>/` — бизнес-логика конкретной фичи: `repository/`, `service/`, `transport/`, композиционный `module.go`. Сейчас две фичи: `shortener` (создание/редирект ссылок, продюсер кликов) и `analytics` (consumer кликов, сохранение в БД).
- Не смешивать: `core/` не должен знать о `ClickEvent`, `Link` и т.д. как о специфике фичи — только общий контракт `events.ClickEvent`. **Это правило сейчас нарушено** — см. `docs/AI_HANDOFF.md`, раздел "Известные баги".

### Паттерн "интерфейс + драйвер" для каждой внешней технологии

Для Postgres, Redis и Kafka используется одна и та же форма — интерфейс с сентинел-ошибками на одном уровне, конкретная реализация под конкретную библиотеку в подпакете:

| Технология | Интерфейс/ошибки | Драйвер |
|---|---|---|
| Postgres | `core/repository/postgres/pool` (`pool.Pool`, `pool.Rows`, `pool.Row`, `pool.CommandTag`, `pool.ErrNoRows`/`ErrViolatesForeignKey`/`ErrUniqueViolation`/`ErrUnknown`) | `core/repository/postgres/pool/pgx` (`pgx.Pool`, оборачивает `jackc/pgx/v5`) |
| Redis | `core/repository/redis` (`cache.Pool`, `cache.StringCmd`, `cache.StatusCmd`, `cache.ErrNotFound`) | `core/repository/redis/goredis` (`goredis.Redis`, оборачивает `redis/go-redis/v9`) |
| Kafka | `core/messaging/gokafka` (`gokafka.Writer`, `gokafka.Reader`, `gokafka.Message`) | `core/messaging/gokafka/segmentio` (`segmentio.Writer`/`segmentio.Reader`, оборачивает `segmentio/kafka-go`) |

Пакет для Kafka называется `gokafka`, не `kafka` — специально, чтобы не конфликтовать с идентификатором `kafka`, под которым по умолчанию импортируется сама библиотека `segmentio/kafka-go` в файлах драйвера (`segmentio/reader.go` импортирует и то, и другое в одном файле).

Каждый адаптер (`pgxRow`/`pgxRows`/`commandTag` в pgx; `goredisStringCmd`/`goredisStatusCmd` в goredis) переводит ошибки конкретной библиотеки в сентинел-ошибки своего уровня через `mapErrors`. **Все** методы адаптера обязаны идти через `mapErrors` — если добавляешь новый метод обёртки, не забудь его туда завести (однажды уже был баг: `Pool.Query`/`pgxRows` не мапили ошибки, пока не поправили).

Фичи зависят только от интерфейсов (`pool.Pool`, `cache.Pool`, `gokafka.Writer`, `gokafka.Reader`), никогда от конкретных `pgx.Pool`/`goredis.Redis`/`segmentio.Writer` в сигнатурах конструкторов. Единственное место, где конкретные типы собираются — `cmd/urlshortener/main.go`.

### Интерфейсы объявляются там, где потребляются

Пример: `internal/features/shortener/service/service.go` объявляет `Repository`/`ClickRecorder` — интерфейсы, которые нужны сервису, а не импортирует конкретные типы из `repository/postgres` или `recorder`. Аналогично `cached.UnderlyingRepository` в `repository/cached/repository.go` объявлен локально в `cached`, а не заимствован из `service`. `internal/features/analytics/processor/processor.go` объявляет свой `ClickRepository` — тот же паттерн. Избегай ситуации, когда repository-слой импортирует пакет service-слоя ради интерфейса.

### Именование пакетов

- Короткое имя, без подчёркиваний, без префиксов вида `core_`/`shortener_` (были и вычищены).
- Имя пакета обычно совпадает с именем своей директории (`pool`, `pgx`, `goredis`, `segmentio`, `events`, `gokafka`), кроме случаев сознательного расхождения ради смысла или коллизии (директория `redis/` → пакет `cache`, потому что это кеш, а не "вообще redis"; директория `messaging/gokafka/` → пакет `gokafka`, а не `kafka`, чтобы не конфликтовать с импортом самой библиотеки `segmentio/kafka-go`).
- Не допускать stutter (`cache.Cache`, `writer.Writer`) — если тип называется так же, как пакет, переименуй один из них.
- Коллизии с чужими пакетами (`net/http`, сторонние библиотеки с тем же именем) решаются переименованием **своего** пакета в уникальное имя (`shortenerhttp` вместо `http`) — никогда добавлением защитного префикса ко всем пакетам "на всякий случай".
- Акронимы — консистентный регистр: `HTTPServer`, `HTMLResponse`, `HTTPResponseHandler` — никогда `Http`/`Html`.

### Ошибки и HTTP-ответы

- Единый набор сентинел-ошибок в `internal/core/errors` (`apperrors.ErrNotFound`, `ErrInvalidArgument`, `ErrConflict`) плюс структурная `apperrors.ValidationErrors` (`[]FieldError`, оборачивает `ErrInvalidArgument` через `Unwrap()`, так что `errors.Is` продолжает работать).
- `core/transport/http/response.HTTPResponseHandler.ErrorResponse` — единственное место, которое матчит сентинелы на HTTP-статусы и `code` (`not_found`/`invalid_argument`/`conflict`/`internal_error`).
- **Сырой `err.Error()` никогда не попадает в тело ответа клиенту** — только в лог через `zap.Error(err)`. Клиент получает `code` + написанный вручную безопасный `message` + (для валидации) `details` — по одному полю на нарушенное правило.
- Если добавляешь новый путь ошибки — заводи её через `fmt.Errorf("...: %w", apperrors.ErrXxx)`, не изобретай новый формат ответа.

### Контекст (`context.Context`)

- `r.Context()` от `net/http` отменяется, как только хендлер возвращает управление — почти сразу после вызова сервиса. Любая работа, которая должна пережить сам HTTP-запрос (запись клика в Kafka в `recorder.RecordClick`), обязана либо стартовать от `context.Background()`, либо (с `Async: true` в writer'е, см. ниже) полагаться на то, что сам вызов не блокируется. Один раз уже был баг, когда таймаут записи в Kafka случайно завели от входящего `ctx` вместо независимого — писалось нестабильно, потому что запись отменялась вместе с завершением HTTP-запроса.
- Логгер с `request_id` кладётся в контекст в `middleware.Logger` и достаётся через `logger.FromContext(ctx)` — эта функция **паникует**, если логгера в контексте нет. Порядок middleware в `cmd/urlshortener/main.go` (`CORS, RequestID, Logger, Trace, Panic`) важен: `Logger` обязан отработать раньше `Trace`/`Panic`, иначе паника при каждом запросе. Не переставлять без явной причины.
- Долгоживущие фоновые компоненты, не привязанные к HTTP-запросу (Kafka producer/consumer — `segmentio.Writer`, `processor.ClickProcessor`), получают логгер **через конструктор** (`log *logger.Logger` параметром), а не через `logger.FromContext(ctx)` — у них нет и не может быть `request_id`, ради которого этот механизм существует. Раньше `ClickProcessor.Start(ctx)` вызывал `logger.FromContext(ctx)` на общем контексте приложения (без `WithLogger`) и падал паникой при старте — пофикшено переходом на конструкторную инъекцию, не откатывай обратно на `FromContext` для фоновых компонентов.

### Не абстрагировать заранее

Абстракция/обобщение вводится, когда появляется второй реальный потребитель, а не "на будущее". В этом проекте так уже случалось несколько раз в обе стороны: убрали `ApiVersionRouter`, когда версия API всего одна; добавили `AsyncWriter`-обёртку с пулом горутин вокруг Kafka-writer, затем убрали её, когда `Async: true` сделал её бесполезной. Если предлагаешь общий механизм "на будущее" (например, обобщить `processor`/consumer-цикл в `core/messaging/gokafka`, пока в проекте один consumer) — сначала спроси, не жди второго потребителя молча, но и не добавляй абстракцию по умолчанию.

### Проверять, а не верить на слово

Самые серьёзные баги и решения в этом проекте нашлись чтением реальных исходников используемых библиотек (`segmentio/kafka-go`) и живым прогоном (`docker exec`, `kafka-console-consumer`, реальные HTTP-запросы) — не рассуждениями о том, как "должно быть", и не советами "прочитал в интернете". Если предлагаешь изменение, основанное на памяти/интуиции о поведении библиотеки — по возможности проверь по исходникам или запуском, а не полагайся на то, что кажется логичным.

### Конфигурация

- Каждый пакет, которому нужна конфигурация из env, имеет свой `Config` + `NewConfig() (Config, error)` + `NewConfigMust() Config` (паникует). Единообразно во всех пакетах (`logger`, `pgx`, `goredis`, `server`, `cached`, `segmentio`, `recorder`, `processor`) — не отступай от этой формы и не называй иначе.
- Топик Kafka (`KAFKA_TOPIC`) сознательно шарится между продюсером (`recorder.Config`) и консьюмером (`processor.Config`) — оба читают один и тот же env-неймспейс `KAFKA_*`, чтобы топик не мог разъехаться между сторонами. А вот `KAFKA_BATCH_SIZE`/`KAFKA_BATCH_TIMEOUT` тоже сейчас шарятся между ними, хотя концептуально это разные настройки (batching сетевой отправки у writer'а vs batching вставки в БД у processor'а) — известное ограничение, не путать с "топик шарить нормально".
- `.env` — реальный локальный файл, в `.gitignore`. `.env.example` — трекается, должен всегда содержать те же ключи, что и `.env` (несколько раз ловили расхождение — переменная добавлена в один файл, забыта в другом, из-за чего приложение падает на старте с `required key ... missing`).

## Команды

```bash
make run                    # go mod tidy && go run cmd/urlshortener/main.go
make env-up                 # поднять docker-compose (postgres, redis, kafka)
make env-down                # остановить всё
make migrate-up             # применить миграции Postgres (golang-migrate)
make migrate-down           # откатить миграции
make migrate-create name=X  # создать новую пару миграций
make kafka-topic-init       # создать топик Kafka идемпотентно (docker-compose exec kafka ..., НЕ через docker-compose run — advertised.listeners настроен под доступ с хоста, а не из соседнего контейнера, см. docs/AI_HANDOFF.md)
make env-port-forward       # прокинуть порт Postgres наружу (socat, профиль tools)

go build ./...              # сборка — СЕЙЧАС ПАДАЕТ, см. docs/AI_HANDOFF.md
go vet ./...                # статический анализ
go test ./...                # тестов нет, но команда должна отрабатывать без ошибок компиляции
gofmt -l .                   # ⚠ покажет "неотформатированными" почти все файлы — это CRLF (git core.autocrlf=true на Windows), НЕ реальные проблемы форматирования. Не запускай gofmt -w . по всему репозиторию, это создаст огромный шумный diff из-за перевода строк.
```

## Ограничения и вещи, которые нельзя менять без явной необходимости

- Не переводить `Async: true` обратно в синхронный режим Kafka-writer и не убирать `Completion`-колбэк в `segmentio/writer.go` (продюсер) — осознанный выбор: `WriteMessages` не блокирует редирект, ценой того, что внутренняя очередь `kafka-go` в этом режиме ничем не ограничена (обычный `append`, без cap) и при затяжном сбое брокера может расти неограниченно. Разбирали подробно и сознательно выбрали "не терять клики" вместо "ограниченная память ценой дропа" — не пытайся "починить" добавлением своего пула горутин/канала без явного запроса, эта тема уже поднималась и отклонялась.
- Не переносить логику `RecordClick` обратно на входящий HTTP-контекст (см. раздел "Контекст" выше).
- Не добавлять префиксы `core_`/`feature_` к именам пакетов "для ясности" — эта практика была явно отвергнута в пользу коротких имён + переименования при коллизии.
- Не запускать `gofmt -w` или любой другой инструмент массового форматирования по всему репозиторию — CRLF-noise, см. выше.
- Тестов нет — это известный, осознанный пробел, а не то, что нужно молча оставлять: если делаешь существенное изменение в бизнес-логике (`service/`, `recorder/`, `processor/`), стоит поднять вопрос о тестах явно, а не считать их необязательными навсегда.
- `migrations/000001_init.down.sql` — уже содержит правильное имя схемы (`urlshortener`), исторически там была опечатка (`todoapp` от другого проекта) — если увидишь `todoapp` снова, это регресс, не оригинальное поведение.
- Короткий код ссылки — `crypto/rand`, не `math/rand` (осознанно, чтобы ссылки не угадывались) — см. `internal/features/shortener/service/create_short_link.go`. Длина 7 символов, колонка `short_code VARCHAR(10)` — есть запас, но не увеличивай длину кода выше 10 без миграции.
- `events.ClickEvent.ID` — `uuid.UUID`, генерируется на продюсере (`events.NewClickEvent`), используется как PRIMARY KEY в `urlshortener.clicks` и как ключ идемпотентности через `ON CONFLICT (id) DO NOTHING` в `analytics/repository/postgres/save_clicks.go` — не убирай это поле и не меняй его тип без пересмотра идемпотентности consumer'а.
