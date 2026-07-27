package analytics

import (
	"github.com/vladislav-koval/url-shortener/internal/core/logger"
	"github.com/vladislav-koval/url-shortener/internal/core/messaging/gokafka"
	"github.com/vladislav-koval/url-shortener/internal/core/repository/postgres/pool"
	"github.com/vladislav-koval/url-shortener/internal/features/analytics/consumer"
	"github.com/vladislav-koval/url-shortener/internal/features/analytics/repository/postgres"
)

type Module struct {
	Consumer *consumer.ClickConsumer
}

// NewModule собирает консьюмера, но не запускает его: количеством горутин и их
// жизненным циклом владеет точка сборки (main.go) — только она знает, когда
// можно закрывать пул Postgres и reader. Конфиг приходит снаружи по той же
// причине, по которой сюда не передаётся ctx: модуль не решает, как его
// запускать, и не читает окружение сам.
func NewModule(pool pool.Pool, reader gokafka.Reader, log *logger.Logger, cfg consumer.Config) *Module {
	repository := postgres.NewRepository(pool)

	return &Module{
		Consumer: consumer.NewClickConsumer(reader, repository, log, cfg),
	}
}
