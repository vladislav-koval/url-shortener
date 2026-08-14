# CLAUDE.md

Стабильные правила и знания о проекте `url-shortener`. Если что-то здесь противоречит коду — доверяй коду и поправь этот файл.

## Что это за проект

Укорачиватель ссылок на Go, разделённый на 2 независимых сервиса (2 бинарника, общий Go-модуль и общий `internal/`):

- **`cmd/urlshortener`** (`internal/shortener/{urls,auth,stats}`) — создание короткого кода, редирект, кеш в Redis, Kafka-паблишинг кликов, Google-логин, `GET /clicks` (своя статистика: короткие коды из своей БД + счётчики кликов по grpc из `analytics`).
- **`cmd/analytics`** (`internal/analytics/{clicks,stats}`) — консьюмит клики из Kafka в Postgres, отдаёт агрегаты по grpc (`AnalyticsService.GetLinkClickCounts`) для `urlshortener`.

Оба процесса запускаются локально (`make run` / `make run-analytics`) и подключаются к общей инфраструктуре из `docker-compose.yml` (Postgres/Redis/Redpanda/migrate/port-forwarder); сами `urlshortener`/`analytics` в `docker-compose.yml` пока не описаны как отдельные сервисы/контейнеры. **У `redis` в `docker-compose.yml` нет volume вообще** — кеш ссылок и сессии переживают рестарты только потому, что сам контейнер ни разу не пересоздавался; `docker rm`/полный `down` уничтожит всё без предупреждения, в отличие от Postgres/Redpanda с именованными volume'ами (`env-cleanup` их и спрашивает отдельно).

Учебный/pet-проект одного разработчика, не production-нагрузка, часто разворачивается на дешёвом сервере. Kafka-writer **не** использует `Async: true`/неограниченную очередь `kafka-go` — вместо этого свой батчинг поверх ограниченной очереди (`QueueSize`, `segmentio.Writer`). Осознанный компромисс в пользу "лучше потерять клик, чем уронить процесс OOM'ом" — на такой машине неограниченная очередь при недоступном брокере реальный риск, не гипотетический. Потери не безмолвные: `Writer.Dropped()` считает и переполнение очереди, и ошибки отправки в Kafka, фоновая горутина раз в интервал (не на каждый дроп — лог-шторм под нагрузкой) сбрасывает дельту в лог. Не считай ни ограниченную очередь, ни дропы багом — это документированное решение, см. "Ограничения".

## Стек

- Go 1.26.1
- Postgres 18 (`jackc/pgx/v5`) — `urlshortener.links`/`urlshortener.users` (shortener) и `analytics.clicks` (analytics), отдельная схема на сервис в одной физической БД — чтобы по ошибке не залезть в чужую таблицу с другой стороны grpc-границы. FK между сервисами нет и не будет: `clicks.short_code` сознательно не `REFERENCES` на `links` (иначе Kafka-consumer в analytics не смог бы писать клики без сихронного похода в чужую схему)
- Redis 8 (`redis/go-redis/v9`) — кеш поверх Postgres для ссылок
- Redpanda (single-node, Kafka wire-протокол), `segmentio/kafka-go` как клиент — приложение говорит с ним как с обычным Kafka-брокером, драйвер/интерфейсы (`gokafka`) не знают, что за брокер на другом конце. Сменили с реального Kafka ради footprint'а на маленьком VPS (без JVM). Топик создаётся с 3 партициями через `rpk topic create` (`make kafka-topic-init`, `Makefile`)
- `go.uber.org/zap`, `kelseyhightower/envconfig`, `go-playground/validator/v10`, `golang-migrate`
- Тесты: `stretchr/testify` (ассерты) + `go.uber.org/mock`/`mockgen` (моки). Юнитами покрыты `shortener/urls/service`, `shortener/urls/repository/postgres`, `shortener/urls/repository/cached`, `platform/repository/postgres/pool/pgx`, `platform/repository/redis/goredis`, `platform/messaging/gokafka/segmentio`, `analytics/stats/{service,repository/postgres,transport/grpc}`, `shortener/stats/{service,repository/postgres,client}` (grpc-клиент замокан reflect-режимом на `analyticsv1.AnalyticsServiceClient`, без реального сервера), `analytics/clicks/repository/postgres` (`SaveClicks` — включая regression-тест на реальный инцидент: `numFields` рассинхронизировался с числом колонок INSERT при добавлении `country`/`city`, каждый клик падал с "INSERT has more target columns than expressions" и бесконечно передоставлялся; тест проверяет, что плейсхолдеры `$1..$n` в запросе — сплошные, без пропусков/повторов, и их число совпадает с числом переданных `args`) — см. раздел "Тестирование" про паттерны. Не покрыты: `platform/transport/http/response`, `shortener/urls/domain` (только косвенно через `service`), `platform/apperrors`, `analytics/clicks/{clicks,consumer}`, весь HTTP-транспорт (`shortener/urls/transport/shortenerhttp`, `shortener/stats/transport/statshttp`, `shortener/auth/transport/authhttp`), `shortener/urls/producer`, голый интерфейс `platform/messaging/gokafka`, `platform/shutdown`, `platform/transport/grpc/{server,client}` (оркестрация с реальными сетевыми примитивами — территория `testcontainers-go`, ещё не подключён). CI (`.github/workflows/test.yml`) — 2 джобы: `proto-check` (`make tools` → `make proto-lint` → `make proto-check`) блокирует `run-tests` через `needs:`; `go test -v -coverprofile=... ./...` на push/PR в `main`/`master`, отчёт покрытия в step summary, моки исключены из подсчёта. БД/Redis/Kafka не поднимает — тесты, требующие реальной инфры, там сломаются, если их добавить без `testcontainers-go` (см. "Тестирование"). Линтера на Go-код кроме `go vet`/`gofmt` нет — на `.proto` есть `easyp lint` (см. "grpc-граница" ниже).

## Архитектурные правила

### `internal/platform/` vs фичи-сервисы (`internal/shortener/`, `internal/analytics/`)

`platform/` не знает о бизнес-логике фичи (технология: Postgres/Redis/Kafka, HTTP-транспорт, общие ошибки `platform/apperrors`, общий контракт `platform/messaging/events`). Каждая фича верхнего уровня — это одновременно точка сборки отдельного бинарника: `internal/shortener/` (свой `cmd/urlshortener/main.go`) и `internal/analytics/` (свой `cmd/analytics/main.go`).

Бизнес-логика внутри фичи-сервиса лежит в подпакете, а не прямо в корне — `internal/shortener/urls/` (`repository/`, `service/`, `transport/`, `producer/`, `module.go`) и `internal/analytics/clicks/` (`repository/`, `consumer/`, `module.go`). Так и было задумано: если бы бизнес-логика `shortener` лежала прямо в `internal/shortener/`, при появлении второй бизнес-фичи в этом же сервисе (например, `auth`) не осталось бы места без stutter. `urls`/`clicks` — конкретные предметные имена, не заглушки.

### "Интерфейс + драйвер" на каждую технологию

| Технология | Интерфейс/ошибки | Драйвер |
|---|---|---|
| Postgres | `platform/repository/postgres/pool` (`pool.Pool`, `ErrNoRows`/`ErrViolatesForeignKey`/`ErrUniqueViolation`/`ErrUnknown`) | `.../pool/pgx` (`pgx.Pool`) |
| Redis | `platform/repository/redis` (`cache.Pool`, `cache.ErrNotFound`) | `.../redis/goredis` (`goredis.Redis`) |
| Kafka | `platform/messaging/gokafka` (`gokafka.Writer`, `gokafka.Reader`, `gokafka.Message`) | `.../gokafka/segmentio` (`segmentio.Writer`/`Reader`) |

Пакет `gokafka`, не `kafka` — чтобы не конфликтовать с идентификатором `kafka` от самой библиотеки `segmentio/kafka-go` в файлах драйвера. Каждый адаптер обязан мапить ошибки через `mapErrors` на **всех** методах без исключений. Фичи зависят только от интерфейсов, конкретные типы собираются только в `cmd/urlshortener/main.go` и `cmd/analytics/main.go` — каждый бинарник свой.

### Интерфейсы объявляются там, где потребляются

`service.Repository`/`service.ClickRecorder` (`shortener/urls/service`), `cached.UnderlyingRepository` (`shortener/urls/repository/cached`), `consumer.ClickRepository` (`analytics/clicks/consumer`) — каждый объявлен локально там, где используется, не заимствован из чужого пакета.

### Нейминг

Короткие имена пакетов, без подчёркиваний/префиксов (`core_`/`feature_` — отвергнуто). Без stutter (`type == package name` — плохо, отсюда `shortener/urls`, а не `shortener/shortener`). Совпадает с именем директории, кроме сознательных расхождений (`redis/` → `cache`; `messaging/gokafka/` → `gokafka`, не `kafka`). Коллизии с чужими пакетами решаются переименованием **своего** пакета (`shortenerhttp` вместо `http`), не защитным префиксом "на всякий случай". Kafka-специфичные имена предпочтительны родовым (`producer`/`consumer`, не `provider`/`processor` — точные Kafka-термины). Акронимы — консистентный регистр (`HTTPServer`, никогда `Http`).

### Ошибки → HTTP

Сентинелы в `platform/apperrors` (`ErrNotFound`, `ErrInvalidArgument`, `ErrConflict`, `ErrAuthorization`, `ErrUnauthenticated`) + `apperrors.ValidationErrors`. `platform/transport/http/response.HTTPResponseHandler.ErrorResponse` — единственное место, матчащее сентинелы на HTTP-статус и `code`. Сырой `err.Error()` никогда не уходит клиенту, только в лог. Актуально для `cmd/urlshortener`; у `cmd/analytics` HTTP-транспорта нет вовсе, только grpc.

`links.user_id REFERENCES users.id ON DELETE SET NULL` — удаление юзера автоматически анонимизирует его ссылки, `urls`-фиче ничего чинить не надо. Но `INSERT` со старым `user_id` от уже протухшей сессии (юзера снесли/накатили старый дамп, а Redis-сессия ещё жива) даёт `pool.ErrViolatesForeignKey`, замаплен в `create_short_link.go` на `apperrors.ErrUnauthenticated` (401), не `ErrConflict` — с точки зрения клиента это не конфликт данных, а невалидная идентификация. Кука на этом пути сама не чистится (автоочистка в `middleware.CurrentUser` срабатывает только когда **сессии** нет в Redis, а тут сессия жива) — не баг, TTL сам разрулит; активно чистить не стали — событие только от ручного вмешательства в БД (дамп/restore), самим приложением недостижимо.

### grpc-граница между `urlshortener` и `analytics`

`platform/transport/grpc/{server,client,interceptor}` — общий слой. `api/proto` → `api/gen` (`easyp`, `make proto`), сгенерённый код **закоммичен**, не в `.gitignore` — иначе `git diff`/`git status` не видят untracked-путь и `make proto-check` (сравнивает свежую генерацию с закоммиченной) молча всегда зелёный, не ловит рассинхрон `.proto` ↔ `api/gen`.

`easyp.yaml`: правило `PACKAGE_DIRECTORY_MATCH` сознательно выключено. В самом `easyp` (проверено v0.16.6 и v0.16.7-rc1) `Root` для этого правила захардкожен в `"."` (`internal/rules/builder.go`, `// TODO: fix me` в исходниках библиотеки), не настраивается ни через `easyp.yaml`, ни через env — правило гарантированно падает на любом `.proto` не в корне репозитория. Не переименовывать `package` под этот баг (сломает wire-имя grpc-сервиса, `/analytics.v1.AnalyticsService/...` в `_FullMethodName`) — просто держать правило выключенным.

Порядок grpc-интерцепторов в `cmd/analytics/main.go` важен: `Validation → Logger → Error → Logging → Panic`. `interceptor.Logger` кладёт логгер в `ctx`; всё, что читает его через `logger.FromContext(ctx)` (`Error`, `Logging`), обязано идти **после** него в списке — `ctx` в цепочке интерцепторов течёт от внешнего к внутреннему и не течёт обратно, более ранний интерцептор никогда не увидит того, что положил в `ctx` более поздний. `interceptor.Logging()` — общий адаптер под `logging.UnaryClientInterceptor`/`UnaryServerInterceptor` (`go-grpc-middleware`), тоже читает логгер из `ctx`, не из захваченного при конструировании — иначе `request_id` не попадает в grpc-логи клиента.

`grpcclient.NewGRPCClient` зовёт `conn.Connect()` сразу при создании, не блокируясь. `grpc.NewClient` сам по себе ленивый — коннект не начнётся до первого RPC, а на Windows резолв `localhost` иногда съедает весь ретрай-бюджет (`GRPC_CLIENT_MAX_RETRIES`×`GRPC_CLIENT_PER_TIMEOUT`) прямо на первом запросе. Не заменять на блокирующее ожидание `READY` при старте процесса — `urlshortener` должен подниматься, даже если `analytics` ещё не готов, и самовосстанавливаться через штатный реконнект `grpc.ClientConn`, когда `analytics` появится.

### Контекст

`r.Context()` умирает вместе с HTTP-запросом — фоновая работа (Kafka-запись) стартует от `context.Background()`. Долгоживущие фоновые компоненты (Kafka producer/consumer) получают логгер **через конструктор** (`log *logger.Logger`), не через `logger.FromContext(ctx)` — у них нет `request_id`, и раньше это уже роняло процесс паникой. Middleware-порядок в `cmd/urlshortener/main.go` (`CORS, RequestID, Logger, Trace, Panic`) — `Logger` обязан быть раньше `Trace`/`Panic`. `cmd/analytics` этой цепочки не имеет — там нет HTTP-запросов, только Kafka-consumer'ы с логгером через конструктор.

### Конфигурация

Каждый пакет — свой `Config` + `NewConfig()`/`NewConfigMust()`, через `envconfig`. **Топик Kafka читается один раз для обеих сторон**: и `producer.Config` (`shortener/urls/producer`), и `consumer.Config` (`analytics/clicks/consumer`) внутри `NewConfig()` отдельным вызовом `envconfig.Process("KAFKA", &topicConfig{...})` берут общий `KAFKA_TOPIC` — так он не может разъехаться между продюсером и консьюмером. Остальные настройки — свои префиксы: `KAFKA_PRODUCER_BATCH_SIZE`/`KAFKA_PRODUCER_BATCH_TIMEOUT` (producer), `KAFKA_CONSUMER_BATCH_SIZE`/`KAFKA_CONSUMER_BATCH_TIMEOUT`/`KAFKA_CONSUMER_GOROUTINES_COUNT` (consumer). Не сливай их обратно в общий `KAFKA_BATCH_SIZE` и не заводи `KAFKA_CONSUMER_TOPIC` отдельной переменной — то и другое уже было и намеренно исправлено.

Сейчас оба процесса запускаются из одного `.env` в корне репозитория (`Makefile`: `include .env` / `export`, общий и для `run`, и для `run-analytics`), поэтому `KAFKA_TOPIC` физически не может разъехаться — он один файл на оба бинарника. Если когда-нибудь `urlshortener`/`analytics` получат раздельные `.env`/переменные окружения (отдельные контейнеры, разный деплой) — это единственное место, где разъезд станет возможен снова, и его придётся страховать иначе (не просто общим `envconfig.Process("KAFKA", ...)`).

`.env` и `.env.example` должны содержать одинаковый набор ключей.

### Остановка: `Run()` + `Shutdown(ctx)`, бюджет один на процесс, оркестрация — общая

Долгоживущие компоненты (`server.HTTPServer`, `consumer.ClickConsumer`, `segmentio.Writer`) имеют пару методов: `Run()` работает, `Shutdown(ctx)` останавливает. Контекста в `Run` нет намеренно — отмена умеет только "брось немедленно", а на выключении надо успеть дослать батч; форма взята у `http.Server` (`ListenAndServe`/`Shutdown(ctx)`).

**Ни один компонент не заводит собственный `context.WithTimeout` на остановку.** Такой контекст в проекте ровно один на процесс, из `SHUTDOWN_TIMEOUT`, и передаётся параметром в каждый `Shutdown` — поэтому фаза, начавшаяся позже (дренаж writer'а), получает остаток, а не свои полные 15с. Не возвращать `ShutdownTimeout` в конфиги фич.

**Оркестрация запуска/остановки — общая функция `platform/shutdown.Run(ctx, log, timeout, runners, unblock, closers...)`, не копипаста в каждом `main.go`.** Раньше (до разделения на 2 бинарника) это был один `main.go`, и весь этот код жил там же. После разделения он на короткое время оказался продублирован один в один (~40 строк: `errgroup` + `stopped`-канал-барьер + `select` на барьер/бюджет + одинаковые лог-сообщения) в `cmd/urlshortener/main.go` и `cmd/analytics/main.go` — вынесен в `platform/shutdown`, как только реально появился второй потребитель (см. "не абстрагируй заранее" в "Ограничениях" — это тот самый случай, когда обобщение уже оправдано, а не заранее).

Форма вызова: `runners []func() error` — то, что уходит в `errgroup.Go`; один `unblock func(context.Context)` — вызов, который заставляет заблокированные `Run()` реально вернуться (`httpServer.Shutdown` в `urlshortener`, `Consumer.Shutdown` на каждом консьюмере в `analytics`); `closers ...func(context.Context)` — вызываются по порядку **только если** все `runners` вернулись в рамках бюджета. Если бюджет исчерпан — `closers` не вызываются вовсе, функция возвращается сразу: ресурс мог всё ещё использоваться зависшим компонентом, закрывать его в этот момент значит гонку или панику. Специфичные лог-сообщения по каждому ресурсу ("failed to shutdown http server" и т.п.) остаются в замыканиях `closers`/`unblock`, определённых в `main.go` — сама `shutdown.Run` не знает, что именно закрывает.

Триггер остановки у каждого компонента свой, под то, что уже даёт его примитив:
- `segmentio.Writer` — закрываемый канал `stop` + `sync.Once`; бюджет доезжает до фоновой горутины полем `drainCtx`, присвоенным строго ДО `close(stop)` — закрытие канала synchronized-before приёма, который из-за него вернулся, поэтому мьютекс не нужен.
- `consumer.ClickConsumer` — **один инстанс на одну горутину/один `Reader`** (см. следующий раздел, почему), у каждого свой независимый `context.WithCancel`; `Shutdown` пишет поле `shutdownCtx`, потом зовёт `cancelFetch()`. При `GoroutinesCount` консьюмеров в `cmd/analytics/main.go` `unblock` вызывает `Shutdown` на каждом по очереди.
- `server.HTTPServer` — отдельного триггера нет: `http.Server.Shutdown(ctx)` уже сам блокируется до выхода `ListenAndServe`, `Run()` просто возвращает управление, когда тот вернулся.

Интерфейс `gokafka.Writer` содержит только `AsyncWriteMessage` — методов жизненного цикла в контрактах для фич быть не должно, ими владеет точка сборки, у которой на руках конкретный тип.

## Тестирование

### Мокать можно только то, что стоит за интерфейсом

Перед тем как пытаться подменить тип в тесте — проверь `go doc <pkg>.<Type>`, интерфейс это или конкретная структура. Только интерфейс можно замокать.

- Свои интерфейсы (`service.Repository`, `pool.Pool`, `cache.Pool`, `cached.UnderlyingRepository`) — `mockgen` в **source-режиме**, `//go:generate mockgen -source=./X.go -destination=mocks/mock_X.go -package=mocks` прямо над объявлением интерфейса, мок — в подпакет `mocks/` рядом.
- Чужие интерфейсы, которые мы не объявляем (`pgx.Row`, `pgx.Rows` — они в pgx намеренно сделаны интерфейсами "to allow tests to mock") — `mockgen` в **reflect-режиме**: `mockgen -destination=mocks/mock_X.go -package=mocks <import-path> Iface1,Iface2` (без `-source`), директива — рядом с обёрткой, которая их использует (`platform/repository/postgres/pool/pgx/adapters.go`).
- Чужие конкретные структуры (`redis.StringCmd`, `redis.StatusCmd` из `go-redis` — НЕ интерфейсы) — мокать нечем и не через что. Если библиотека сама даёт тестовые конструкторы (`redis.NewStringResult(val, err)`, `redis.NewStatusResult(val, err)`) — использовать их, `mockgen` тут бесполезен.
- `*pgxpool.Pool` и свободные функции пакета (`pgxpool.ParseConfig`, `pgxpool.NewWithConfig`) — не юнит-тестируемы в принципе: у Go нет точки подмены для функций пакета (не методов интерфейса), а `*pgxpool.Pool` — конкретный тип без замены. Это осознанно территория интеграционных тестов (`testcontainers-go`, ещё не подключён — следующий шаг; CI на `ubuntu-latest` уже даёт Docker без доп. настройки).

### Тест на `mapErrors` — всегда в два слоя

`mapErrors` есть в каждом адаптере (`pool/pgx/adapters.go`, `redis/goredis/adapters.go` — оба под `platform/repository/...`) и мапит ошибки библиотеки в свои сентинелы. Один тест на голую функцию **не ловит** регресс "кто-то забыл вызвать `mapErrors` внутри метода-обёртки и оставил `return err`" — сигнатуры совпадают, компилируется, просто ошибка перестаёт матчиться выше по стеку. Поэтому два слоя:
1. `mapErrors(err)` напрямую на всех ветках — чистая функция, без моков.
2. Через саму обёртку (`pgxRow{mockRow}.Scan(...)`, `goredisStringCmd{realCmd}.Result()`) — доказывает, что обёртка её реально вызывает.

Regression-тест на реальный инцидент этого проекта: локальная БД без применённых миграций → `relation "urlshortener.links" does not exist` (Postgres-код `42P01`) → не совпадает ни с `pgx.ErrNoRows`, ни с `23503`/`23505` → падает в generic-ветку → `500` вместо `404`. Зафиксировано тестом `*pgconn.PgError{Code: "42P01"} → pool.ErrUnknown` — не удалять.

### Паттерны тестового кода

- `initTest(t *testing.T) *testFixture` — общий сетап через структуру с именованными полями, не позиционный возврат 5+ значений (см. `shortener/urls/repository/cached/repository_test.go`). Обязательно `t.Helper()` внутри.
- `gomock.NewController(t)` (`go.uber.org/mock`, не устаревший `golang/mock`) сам регистрирует `t.Cleanup(ctrl.Finish)` — не нужен ручной `teardown`.
- `EXPECT()` без `.Return(...)` (и `.Return(nil)` тоже) на методе с интерфейсным типом возврата даёт `nil`-интерфейс — вызов любого метода на нём паникует ("invalid memory address"). Нужен настоящий (пусть и поддельный) объект, а не `nil`.
- Табличные тесты (`[]struct{name string; ...; check func(t, ...)}` + `t.Run`) — когда кейсы совпадают по форме arrange→act→assert. Когда ассерты кейсов принципиально разные — поле `check func(t *testing.T, ...)` вместо разрастающихся `wantX`/`wantY`. Кейсов мало (2-3) и форма мока разная — отдельные функции нормальны, не насиловать таблицей.
- Группировка через `t.Run` (аналог `describe`/`it`) — можно мешать цикл по таблице и ручные `t.Run(...)` в одной родительской `TestXxx`. Не заводить лишний уровень вложенности, если он не несёт смысла (группировка "эти из цикла, эти нет" — не смысловая).
- `logger.FromContext(ctx)` паникует без логгера в контексте (см. "Контекст" выше) — HTTP-путь читает логгер из ctx, тесты обязаны положить его сами: `logger.WithLogger(ctx, &logger.Logger{Logger: zap.NewNop()})` (тихий) или `zap.New(core)` с `core, logs := observer.New(zap.ErrorLevel)` из `go.uber.org/zap/zaptest/observer`, если нужно проверить сам факт/текст лога.
- Не дублировать один и тот же литерал между `EXPECT()`-матчером и реально переданным аргументом вручную (могут разъехаться при правке одного места) — заводить `const input... = "..."` один раз на файл/функцию.

## Команды

```bash
make run                    # go mod tidy && go run cmd/urlshortener/main.go
make run-analytics          # go mod tidy && go run cmd/analytics/main.go
make env-up / env-down      # docker-compose (postgres, redis, redpanda)
make migrate-up / migrate-down
make kafka-topic-init       # docker-compose exec redpanda rpk topic create ... (НЕ run — advertise-kafka-addr настроен под доступ с хоста)
make tools                  # установить easyp + protoc-gen-go/-go-grpc/-validate, зафиксированных версий
make proto                  # easyp mod download && easyp generate — api/proto -> api/gen
make proto-lint             # easyp lint
make proto-check            # make proto + git status --porcelain api/gen (ловит и modified, и новые untracked-файлы)
go build ./... && go vet ./...
gofmt -l .                   # покажет почти все файлы как "неформатированные" — это CRLF на Windows, не запускай gofmt -w . по всему репо
```

## Ограничения

- **`Async: true` у Kafka-writer убран намеренно** (в более ранней версии было наоборот, см. "Что это за проект"). Вместо него `segmentio.Writer` сам батчит сообщения через ограниченный канал `queue chan kafka.Message` (`QueueSize`) в фоновой `run()`; `AsyncWriteMessage` — неблокирующий `select`+`default`, `Shutdown(ctx)` закрывает канал-сигнал `stop` и дожидается финального `flush()` через `wg.Wait()`. Сам `queue` при этом НЕ закрывается никогда — у него много конкурентных отправителей, и закрытие дало бы панику "send on closed channel". Причина смены — реальный риск OOM неограниченной очереди на дешёвом сервере, не гипотетический. Не откатывать обратно без явного запроса.
- **Потери из ограниченной очереди обязаны быть видимыми, не откатывать этот механизм.** `Writer.dropped` (atomic-счётчик) инкрементится в двух местах: при переполнении очереди (`AsyncWriteMessage`, ветка `default`) и при ошибке `WriteMessages` внутри `run()`/`flush()` — нужны оба случая, не один. Отдельная горутина `reportDropped()`, запущенная в `NewWriter`, раз в `dropReportInterval` (30с) логирует **дельту** через `Dropped()`, не каждое сообщение по отдельности — под нагрузкой, когда дропы вероятнее всего, лог на каждый дроп сам добавил бы нагрузку в момент деградации. `Shutdown(ctx)` дополнительно логирует итоговую сумму, если она не нулевая.
- `kafka.Writer.BatchTimeout` внутри `segmentio.Writer` захардкожен маленьким (`10ms`) — это не забытая настройка. Батчинг уже делаем сами, `WriteMessages` вызывается с готовым батчем; без маленького `BatchTimeout` внутренний батчер `kafka-go` может ещё подождать, донабирая сообщения по своим правилам (см. доку `Writer.WriteMessages`: "blocks until either a full batch can be assembled or the batch timeout is reached") — лишняя задержка поверх уже собранного батча.
- Не переносить `RecordClick`/Kafka-запись обратно на входящий HTTP-контекст. `RecordClick` → `AsyncWriteMessage` — неблокирующая запись в Go-канал, сети не касается вообще; реальный `WriteMessages` идёт из фоновой `run()`-горутины на `context.Background()`, полностью отвязанной от жизненного цикла HTTP-запроса — поэтому обрыв соединения на `Redirect` (мобильная сеть, закрытая вкладка) никак не может потерять клик через отменённый `ctx`. Поэтому `ClickRecorder.RecordClick(clickEvent events.ClickEvent)` не принимает `ctx` вовсе, а `producer.NewProducer(writer, log)` получает логгер через конструктор, как и `consumer.NewClickConsumer`.
  - Остановка `clickWriter` безопасна даже если где-то ещё идёт `RecordClick`: после `close(stop)` `run()` доделывает последний `flush()` и завершается, `Shutdown(ctx)` дожидается через `wg.Wait()` и только потом закрывает нижележащий `kafka.Writer`.
  - Известный пробел, не Kafka-специфичный: у `http.Server` в `server.go` не выставлены `ReadTimeout`/`WriteTimeout`/`IdleTimeout`/`ReadHeaderTimeout`, и в цепочке middleware нет таймаут-мидлвари — единственная защита от реально зависшего хендлера сейчас это внутренние таймауты `kafka-go` (dial ~5с, read ~10с) и таймаут самого браузера/клиента. Если понадобится жёсткая граница — добавить их на `http.Server` или отдельной мидлварью.
- **Один `gokafka.Reader` на одну горутину консьюмера, не общий reader на несколько горутин.** `analytics/clicks.NewModule` строит по одному `*consumer.ClickConsumer` на каждый переданный `Reader`; `cmd/analytics/main.go` создаёт `KAFKA_CONSUMER_GOROUTINES_COUNT` отдельных `Reader`'ов с одним и тем же `GroupID`, а не один reader на всех. Причина: коммит офсета в Kafka помечает обработанными все более ранние офсеты той же партиции — если бы несколько горутин делили один reader и коммитили свои батчи независимо, быстрая горутина могла закоммитить офсет, обгоняющий ещё не сохранённое в Postgres сообщение медленной горутины из той же партиции, и то сообщение терялось бы молча при падении процесса между этим (баг был реально найден и исправлен, не гипотетический). С отдельным `Reader` на горутину партиции между ними эксклюзивно распределяет сам Kafka group coordinator — офсеты одной партиции коммитит только один reader, и весь путь fetch→save→commit внутри него остаётся строго последовательным. **Число горутин/reader'ов не должно превышать число партиций топика** (сейчас 3, `Makefile` → `kafka-topic-init`) — лишние reader'ы просто останутся без партиций и будут простаивать. Не возвращать общий reader на несколько горутин без явного запроса.
- **Паника в фоновых горутинах не перехватывается нигде, кроме HTTP-хендлеров.** `recover()` в проекте ровно один — в `platform/transport/http/middleware` (`Panic`-мидлварь), и он защищает только горутину, обрабатывающую конкретный HTTP-запрос. `segmentio.Writer.run()`/`reportDropped()` (продюсер, `cmd/urlshortener`) и цикл `analytics/clicks/consumer.ClickConsumer.Run()`/`flush()` (`cmd/analytics`) — долгоживущие фоновые горутины вне жизненного цикла HTTP-запроса, recover через мидлварь до них физически не дотягивается (паника не пересекает границу горутины). Непойманная паника в любой из них уронит весь процесс целиком. Известный, подтверждённый (2026-08-02) пробел, ещё не закрыт — если добавлять `recover()` в эти горутины, делать это осознанно (лог + решение "продолжить работу горутины" vs "дать процессу упасть"), а не как случайную заплатку.
- Не абстрагировать заранее — обобщение только на второй реальный потребитель. Историческая заметка: `AsyncWriter`-обёртка с пулом горутин уже была, её убирали в пользу `Async: true` у `kafka-go`, а потом вернули снова (см. выше) — потому что ограничение памяти на целевом железе оказалось важнее веса абстракции. Тот же принцип сработал и при разделении на 2 бинарника: общая оркестрация запуска/остановки (`platform/shutdown.Run`) вынесена только после того, как появился второй реальный `main.go`, не заранее.
- Проверять по исходникам библиотек и живым прогоном, не по памяти/интуиции — самые серьёзные находки в проекте (Kafka batching, group coordinator not available из-за `offsets.topic.replication.factor=3` на одном брокере, гонка на коммитах офсетов при общем reader'е) нашлись именно так.
- `crypto/rand` для генерации кода ссылки, не `math/rand`. Длина 7, колонка `VARCHAR(10)`.
- `events.ClickEvent.ID` (uuid) — генерируется на продюсере, используется как ключ идемпотентности (`ON CONFLICT (id) DO NOTHING`) в `analytics`. Офсет коммитится только после успешного сохранения в БД (или после `ErrConflict` — это тоже означает "данные уже там", не повод не коммитить).
