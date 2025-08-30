package repo

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	Rdb *redis.Client
}

type Repositories struct {
	Postgres *PostgresRepo
	Redis    *RedisRepo
}

func New(pg *pgxpool.Pool, rdb *redis.Client) *Repositories {
	return &Repositories{
		Postgres: &PostgresRepo{DB: pg},
		Redis:    &RedisRepo{Rdb: rdb},
	}
}

func (r *RedisRepo) Client() *redis.Client {
	return r.Rdb
}
