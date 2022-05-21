package env

import (
	"time"

	"github.com/cockroachdb/errors"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

const (
	// Local 本地开发环境
	Local string = "local"
	// Testing 测试环境
	Testing string = "testing"
	// Staging 预发布环境
	Staging string = "staging"
	// Prod 生产环境
	Prod string = "prod"
)

// Env defines all the environments
// Env 使用 envconfig 去加载，确保应用启动的时候所有的环境变量都有了
type Env struct {
	Tier string `envconfig:"TIER" required:"true" default:"local"`
	Port int    `envconfig:"PORT" required:"true" default:"8080"`

	HTTPReadTimeout  time.Duration `envconfig:"HTTP_READ_TIMEOUT" required:"false" default:"8s"`
	HTTPWriteTimeout time.Duration `envconfig:"HTTP_WRITE_TIMEOUT" required:"false" default:"16s"`
	HTTPIdleTimeout  time.Duration `envconfig:"HTTP_IDLE_TIMEOUT" required:"false" default:"60s"`

	DBSourceName            string        `envconfig:"DB_SOURCE_NAME" required:"true"`
	DBMaxOpenConnections    int           `envconfig:"DB_MAX_OPEN_CONNECTIONS" required:"true"`
	DBMaxIdleConnections    int           `envconfig:"DB_MAX_IDLE_CONNECTIONS" required:"true"`
	DBConnectionMaxLifetime time.Duration `envconfig:"DB_CONNECTION_MAX_LIFETIME" required:"true" default:"1h"`
}

var Envs Env

// Load 读取并解析环境变量到 Envs, 并做一些验证
func Load() error {
	if err := godotenv.Overload("env/.env"); err != nil {
		return errors.Wrap(err, "overload environments failed")
	}
	if err := envconfig.Process("", &Envs); err != nil {
		return errors.Wrap(err, "process environments failed")
	}
	return nil
}
