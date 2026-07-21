# CLAUDE.md

Стабильные правила и знания о проекте `url-shortener`. Если что-то здесь противоречит коду — доверяй коду и поправь этот файл. Восстановлен после случайного удаления; следующей сессии стоит один раз сверить построчно с кодом, а не доверять слепо.

## Что это за проект

Укорачиватель ссылок на Go: создание короткого кода для URL, редирект по короткому коду, кеширование в Redis, публикация событий о переходах в Kafka и их обработка в фиче `analytics` (сохраняет клики в Postgres). Просмотр аналитики по HTTP — ещё не реализован.

Учебный/pet-проект одного разработчика, не production-нагрузка. Kafka-writer работает с `Async: true` и неограниченной внутренней очередью — осознанный компромисс ("не терять клики" важнее "ограниченная память"), не считай это автоматически багом.

## Стек

- Go 1.26.1
- Postgres 18 (`jackc/pgx/v5`) — `urlshortener.links` и `urlshortener.clicks`
- Redis 8 (`redis/go-redis/v9`) — кеш поверх Postgres для ссылок
- Kafka 4.3, KRaft, один брокер (`segmentio/kafka-go`)
- `go.uber.org/zap`, `kelseyhightower/envconfig`, `go-playground/validator/v10`, `golang-migrate`
- Тесты: `stretchr/testify` (ассерты) + `go.uber.org/mock`/`mockgen` (моки). Юнитами покрыты `shortener/service`, `shortener/repository/postgres`, `shortener/repository/cached`, `core/repository/postgres/pool/pgx`, `core/repository/redis/goredis` — см. раздел "Тестирование" про паттерны. Не покрыты: `response/handler.go`, `domain/link.go` (только косвенно через `service`), `apperrors`, весь `analytics/*`, `transport/shortenerhttp`, Kafka-слой (`producer`/`gokafka`). CI (`.github/workflows/test.yml`: `go test -v ./...` на push/PR в `main`/`master`) БД/Redis/Kafka не поднимает — тесты, требующие реальной инфры, там сломаются, если их добавить без `testcontainers-go` (см. "Тестирование"). Линтера кроме `go vet`/`gofmt` нет — в CI линтер не настроен.

## Архитектурные правила

### `internal/core/` vs `internal/features/`

`core/` не знает о бизнес-логике фичи (технология: Postgres/Redis/Kafka, HTTP-транспорт, домен, общие ошибки `core/apperrors`, общий контракт `core/messaging/events`). `features/<name>/` — вся бизнес-логика: `repository/`, `service/`, `transport/`, `module.go`. Фичи: `shortener` (создание/редирект, продюсер кликов — подпакет `shortener/producer`, было `recorder`) и `analytics` (консьюмер кликов — подпакет `analytics/consumer`, было `processor`).

### "Интерфейс + драйвер" на каждую технологию

| Технология | Интерфейс/ошибки | Драйвер |
|---|---|---|
| Postgres | `core/repository/postgres/pool` (`pool.Pool`, `ErrNoRows`/`ErrViolatesForeignKey`/`ErrUniqueViolation`/`ErrUnknown`) | `.../pool/pgx` (`pgx.Pool`) |
| Redis | `core/repository/redis` (`cache.Pool`, `cache.ErrNotFound`) | `.../redis/goredis` (`goredis.Redis`) |
| Kafka | `core/messaging/gokafka` (`gokafka.Writer`, `gokafka.Reader`, `gokafka.Message`) | `.../gokafka/segmentio` (`segmentio.Writer`/`Reader`) |

Пакет `gokafka`, не `kafka` — чтобы не конфликтовать с идентификатором `kafka` от самой библиотеки `segmentio/kafka-go` в файлах драйвера. Каждый адаптер обязан мапить ошибки через `mapErrors` на **всех** методах без исключений. Фичи зависят только от интерфейсов, конкретные типы собираются только в `cmd/urlshortener/main.go`.

### Интерфейсы объявляются там, где потребляются

`service.Repository`/`service.ClickRecorder`, `cached.UnderlyingRepository`, `consumer.ClickRepository` — каждый объявлен локально там, где используется, не заимствован из чужого пакета.

### Нейминг

Короткие имена пакетов, без подчёркиваний/префиксов (`core_`/`feature_` — отвергнуто). Без stutter (`type == package name` — плохо). Совпадает с именем директории, кроме сознательных расхождений (`redis/` → `cache`; `messaging/gokafka/` → `gokafka`, не `kafka`). Коллизии с чужими пакетами решаются переименованием **своего** пакета (`shortenerhttp` вместо `http`), не защитным префиксом "на всякий случай". Kafka-специфичные имена предпочтительны родовым (`producer`/`consumer`, не `provider`/`processor` — точные Kafka-термины). Акронимы — консистентный регистр (`HTTPServer`, никогда `Http`).

### Ошибки → HTTP

Сентинелы в `core/apperrors` (`ErrNotFound`, `ErrInvalidArgument`, `ErrConflict`) + `apperrors.ValidationErrors`. `core/transport/http/response.HTTPResponseHandler.ErrorResponse` — единственное место, матчащее сентинелы на HTTP-статус и `code`. Сырой `err.Error()` никогда не уходит клиенту, только в лог.

### Контекст

`r.Context()` умирает вместе с HTTP-запросом — фоновая работа (Kafka-запись) стартует от `context.Background()`. Долгоживущие фоновые компоненты (Kafka producer/consumer) получают логгер **через конструктор** (`log *logger.Logger`), не через `logger.FromContext(ctx)` — у них нет `request_id`, и раньше это уже роняло процесс паникой. Middleware-порядок в `main.go` (`CORS, RequestID, Logger, Trace, Panic`) — `Logger` обязан быть раньше `Trace`/`Panic`.

### Конфигурация

Каждый пакет — свой `Config` + `NewConfig()`/`NewConfigMust()`, через `envconfig`. **Топик Kafka читается один раз для обеих сторон**: и `producer.Config`, и `consumer.Config` внутри `NewConfig()` отдельным вызовом `envconfig.Process("KAFKA", &topicConfig{...})` берут общий `KAFKA_TOPIC` — так он не может разъехаться между продюсером и консьюмером. Остальные настройки — свои префиксы: `KAFKA_PRODUCER_BATCH_SIZE`/`KAFKA_PRODUCER_BATCH_TIMEOUT` (producer), `KAFKA_CONSUMER_BATCH_SIZE`/`KAFKA_CONSUMER_BATCH_TIMEOUT`/`KAFKA_CONSUMER_GOROUTINES_COUNT` (consumer). Не сливай их обратно в общий `KAFKA_BATCH_SIZE` и не заводи `KAFKA_CONSUMER_TOPIC` отдельной переменной — то и другое уже было и намеренно исправлено.

`.env` и `.env.example` должны содержать одинаковый набор ключей.

## Тестирование

### Мокать можно только то, что стоит за интерфейсом

Перед тем как пытаться подменить тип в тесте — проверь `go doc <pkg>.<Type>`, интерфейс это или конкретная структура. Только интерфейс можно замокать.

- Свои интерфейсы (`service.Repository`, `pool.Pool`, `cache.Pool`, `cached.UnderlyingRepository`) — `mockgen` в **source-режиме**, `//go:generate mockgen -source=./X.go -destination=mocks/mock_X.go -package=mocks` прямо над объявлением интерфейса, мок — в подпакет `mocks/` рядом.
- Чужие интерфейсы, которые мы не объявляем (`pgx.Row`, `pgx.Rows` — они в pgx намеренно сделаны интерфейсами "to allow tests to mock") — `mockgen` в **reflect-режиме**: `mockgen -destination=mocks/mock_X.go -package=mocks <import-path> Iface1,Iface2` (без `-source`), директива — рядом с обёрткой, которая их использует (`pool/pgx/adapters.go`).
- Чужие конкретные структуры (`redis.StringCmd`, `redis.StatusCmd` из `go-redis` — НЕ интерфейсы) — мокать нечем и не через что. Если библиотека сама даёт тестовые конструкторы (`redis.NewStringResult(val, err)`, `redis.NewStatusResult(val, err)`) — использовать их, `mockgen` тут бесполезен.
- `*pgxpool.Pool` и свободные функции пакета (`pgxpool.ParseConfig`, `pgxpool.NewWithConfig`) — не юнит-тестируемы в принципе: у Go нет точки подмены для функций пакета (не методов интерфейса), а `*pgxpool.Pool` — конкретный тип без замены. Это осознанно территория интеграционных тестов (`testcontainers-go`, ещё не подключён — следующий шаг; CI на `ubuntu-latest` уже даёт Docker без доп. настройки).

### Тест на `mapErrors` — всегда в два слоя

`mapErrors` есть в каждом адаптере (`pool/pgx/adapters.go`, `redis/goredis/adapters.go`) и мапит ошибки библиотеки в свои сентинелы. Один тест на голую функцию **не ловит** регресс "кто-то забыл вызвать `mapErrors` внутри метода-обёртки и оставил `return err`" — сигнатуры совпадают, компилируется, просто ошибка перестаёт матчиться выше по стеку. Поэтому два слоя:
1. `mapErrors(err)` напрямую на всех ветках — чистая функция, без моков.
2. Через саму обёртку (`pgxRow{mockRow}.Scan(...)`, `goredisStringCmd{realCmd}.Result()`) — доказывает, что обёртка её реально вызывает.

Regression-тест на реальный инцидент этого проекта: локальная БД без применённых миграций → `relation "urlshortener.links" does not exist` (Postgres-код `42P01`) → не совпадает ни с `pgx.ErrNoRows`, ни с `23503`/`23505` → падает в generic-ветку → `500` вместо `404`. Зафиксировано тестом `*pgconn.PgError{Code: "42P01"} → pool.ErrUnknown` — не удалять.

### Паттерны тестового кода

- `initTest(t *testing.T) *testFixture` — общий сетап через структуру с именованными полями, не позиционный возврат 5+ значений (см. `repository/cached/repository_test.go`). Обязательно `t.Helper()` внутри.
- `gomock.NewController(t)` (`go.uber.org/mock`, не устаревший `golang/mock`) сам регистрирует `t.Cleanup(ctrl.Finish)` — не нужен ручной `teardown`.
- `EXPECT()` без `.Return(...)` (и `.Return(nil)` тоже) на методе с интерфейсным типом возврата даёт `nil`-интерфейс — вызов любого метода на нём паникует ("invalid memory address"). Нужен настоящий (пусть и поддельный) объект, а не `nil`.
- Табличные тесты (`[]struct{name string; ...; check func(t, ...)}` + `t.Run`) — когда кейсы совпадают по форме arrange→act→assert. Когда ассерты кейсов принципиально разные — поле `check func(t *testing.T, ...)` вместо разрастающихся `wantX`/`wantY`. Кейсов мало (2-3) и форма мока разная — отдельные функции нормальны, не насиловать таблицей.
- Группировка через `t.Run` (аналог `describe`/`it`) — можно мешать цикл по таблице и ручные `t.Run(...)` в одной родительской `TestXxx`. Не заводить лишний уровень вложенности, если он не несёт смысла (группировка "эти из цикла, эти нет" — не смысловая).
- `logger.FromContext(ctx)` паникует без логгера в контексте (см. "Контекст" выше) — HTTP-путь читает логгер из ctx, тесты обязаны положить его сами: `logger.WithLogger(ctx, &logger.Logger{Logger: zap.NewNop()})` (тихий) или `zap.New(core)` с `core, logs := observer.New(zap.ErrorLevel)` из `go.uber.org/zap/zaptest/observer`, если нужно проверить сам факт/текст лога.
- Не дублировать один и тот же литерал между `EXPECT()`-матчером и реально переданным аргументом вручную (могут разъехаться при правке одного места) — заводить `const input... = "..."` один раз на файл/функцию.

## Команды

```bash
make run                    # go mod tidy && go run cmd/urlshortener/main.go
make env-up / env-down      # docker-compose (postgres, redis, kafka)
make migrate-up / migrate-down
make kafka-topic-init       # docker-compose exec kafka ... (НЕ run — advertised.listeners настроен под доступ с хоста)
go build ./... && go vet ./...
gofmt -l .                   # покажет почти все файлы как "неформатированные" — это CRLF на Windows, не запускай gofmt -w . по всему репо
```

## Ограничения

- Не откатывать `Async: true` у Kafka-writer и не убирать пул ретраев/`Completion` — см. "Стек".
- Не переносить `RecordClick`/Kafka-запись обратно на входящий HTTP-контекст. Причина не только в стиле: `Redirect` — ровно та точка, где клиент может оборваться (мобильная сеть, закрытая вкладка), и если `WriteMessage` получит уже отменённый `r.Context()`, `Writer.partitions(ctx, topic)` внутри `segmentio/kafka-go` вернёт ошибку ещё до постановки сообщения во внутреннюю очередь — клик потеряется не из-за проблем с Kafka, а из-за поведения браузера, что напрямую противоречит идее `Async: true` ("не терять клики" — см. "Стек"). Само produce-в-брокер уже и так шло через `context.Background()` (см. `Writer.produce` в библиотеке) — отвязка была нужна именно для шага получения метаданных партиции. Поэтому `ClickRecorder.RecordClick(clickEvent events.ClickEvent)` не принимает `ctx` вовсе — `producer.Producer.RecordClick` сам использует `context.Background()` для `WriteMessage`, а `producer.NewProducer(writer, log)` получает логгер через конструктор, как и `consumer.NewClickConsumer` — исправлено.
  - `WriteMessage`/`partitions()` из-за этого **не виснет на 10с на каждый клик** — `Async: true` убирает ожидание ack от брокера, а сам вызов `partitions(ctx, topic)` в штатном режиме читает метаданные из внутреннего кэша `Transport` (`state.metadata`, `MetadataTTL` по умолчанию 6с, обновляется в фоне собственной горутиной транспорта — не связан ни с нашим `ctx`, ни с Async) и возвращается почти мгновенно, без сети. Реальная сетевая блокировка возможна только один раз — на самом первом вызове с момента старта процесса (пока `connPool` не "ready", `transport.go:342`); если брокер недоступен уже после прогрева, `partitions()` вернёт ошибку сразу (`state.err`), а не будет ждать. Так что зависание конкретно из-за Kafka — это только cold-start edge case, не постоянный риск.
  - Закрытие `clickWriter` при shutdown безопасно даже если где-то ещё идёт `RecordClick`: `Writer.Close()` (`writer.go:555`) флашит pending-батчи и ждёт `w.group.Wait()`, а опоздавший `WriteMessages` после `Close()` просто вернёт `io.ErrClosedPipe` (залогируется, без паники). Graceful shutdown HTTP-сервера (`server.Shutdown` с `ShutdownTimeout`, по умолчанию 30с) вообще не зависит от того, какой `ctx` получает `RecordClick` — это два независимых механизма.
  - Известный пробел, не Kafka-специфичный: у `http.Server` в `server.go` не выставлены `ReadTimeout`/`WriteTimeout`/`IdleTimeout`/`ReadHeaderTimeout`, и в цепочке middleware нет таймаут-мидлвари — единственная защита от реально зависшего хендлера сейчас это внутренние таймауты `kafka-go` (dial ~5с, read ~10с) и таймаут самого браузера/клиента. Если понадобится жёсткая граница — добавить их на `http.Server` или отдельной мидлварью.
- Не абстрагировать заранее — обобщение только на второй реальный потребитель (уже добавляли и убирали `AsyncWriter`-обёртку с пулом горутин, когда `Async: true` сделал её лишней).
- Проверять по исходникам библиотек и живым прогоном, не по памяти/интуиции — самые серьёзные находки в проекте (Kafka batching, group coordinator not available из-за `offsets.topic.replication.factor=3` на одном брокере) нашлись именно так.
- `crypto/rand` для генерации кода ссылки, не `math/rand`. Длина 7, колонка `VARCHAR(10)`.
- `events.ClickEvent.ID` (uuid) — генерируется на продюсере, используется как ключ идемпотентности (`ON CONFLICT (id) DO NOTHING`) в `analytics`. Офсет коммитится только после успешного сохранения в БД (или после `ErrConflict` — это тоже означает "данные уже там", не повод не коммитить).
