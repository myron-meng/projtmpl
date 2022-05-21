package database

import (
	"database/sql"

	"projtmpl/env"

	_ "github.com/go-sql-driver/mysql"
)

// NewConnectionPool 建立数据库连接池
// dsn 的格式：${your_username}:${your_password}@tcp(${db_host}:${db_port})/${db_name}
func NewConnectionPool() (*sql.DB, error) {
	dsn := env.Envs.DBSourceName + "?parseTime=true&charset=utf8mb4&loc=UTC"
	maxOpen := env.Envs.DBMaxOpenConnections
	maxIdle := env.Envs.DBMaxIdleConnections
	maxLifetime := env.Envs.DBConnectionMaxLifetime
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(maxLifetime)

	return db, nil
}
