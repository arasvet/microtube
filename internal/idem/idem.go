package idem

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *Service { return &Service{rdb: rdb} }

const (
	valPending = "pending"
	valDone    = "done"
)

type Status int

const (
	Miss      Status = iota // нет ключа, можно обрабатывать (после успешного SETNX)
	Duplicate               // уже есть pending/done -> дубликат
	Unknown                 // ошибка Redis/неопределённо
)

// TryReserve пытается зарезервировать event_id. Возвращает Miss, если это первая обработка.
func (s *Service) TryReserve(ctx context.Context, eventID string) (Status, error) {
	key := "idem:event:" + eventID
	ok, err := s.rdb.SetNX(ctx, key, valPending, 30*time.Second).Result()
	if err != nil {
		return Duplicate, err // на сетевых ошибках лучше перестраховаться и не плодить дубли
	}
	if ok {
		return Miss, nil
	}
	return Duplicate, nil
}

// MarkDone помечает событие как окончательно обработанное на долгий TTL.
func (s *Service) MarkDone(ctx context.Context, eventID string) error {
	key := "idem:event:" + eventID
	return s.rdb.Set(ctx, key, valDone, 48*time.Hour).Err()
}

// Release отменяет резервацию (если БД упала).
func (s *Service) Release(ctx context.Context, eventID string) {
	_ = s.rdb.Del(ctx, "idem:event:"+eventID).Err()
}
