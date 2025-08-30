package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/arasvet/microtube/internal/domain"
	"github.com/arasvet/microtube/internal/idem"
	"github.com/arasvet/microtube/internal/repo"
)

type EventsUC struct {
	store repo.Store
	idem  *idem.Service
}

func NewEventsUC(store repo.Store, idemSvc *idem.Service) *EventsUC {
	return &EventsUC{store: store, idem: idemSvc}
}

type IngestResult struct {
	Inserted bool
}

func (uc *EventsUC) Ingest(ctx context.Context, e domain.Event) (IngestResult, error) {
	// валидируем событие
	if err := e.Validate(); err != nil {
		return IngestResult{}, err
	}

	key := e.EventID.String()
	status, err := uc.idem.TryReserve(ctx, key)
	if err != nil {
		// Redis недоступен — продолжаем, но лог/метрика обязательны
		slog.Warn("idem.TryReserve failed", "err", err, "event_id", key)
		status = idem.Unknown // например, добавить такой статус
	}
	if status == idem.Duplicate {
		// log.Debug("duplicate by idem", "event_id", key)
		return IngestResult{Inserted: false}, nil
	}

	// Управление жизненным циклом ключа
	reserved := status == idem.Miss // мы реально резерв взяли
	committed := false
	// Гарантированный релиз, если не дойдём до MarkDone
	defer func() {
		if reserved && !committed {
			uc.idem.Release(ctx, key)
		}
	}()

	tx, err := uc.store.Begin(ctx)
	if err != nil {
		return IngestResult{}, err
	}
	defer tx.Rollback(ctx)

	// 1) основная запись события
	inserted, err := uc.store.InsertEvent(ctx, tx, e)
	if err != nil {
		return IngestResult{}, err
	}

	if !inserted {
		// дубль: можно сразу финализировать idem-ключ
		if reserved {
			_ = uc.idem.MarkDone(ctx, key)
			committed = true
		}
		return IngestResult{Inserted: false}, nil
	}

	// 2) агрегаты
	if err := uc.store.UpsertVideoCounters(ctx, tx, e); err != nil {
		return IngestResult{}, err
	}
	if err := uc.store.UpsertVideoDaily(ctx, tx, e); err != nil {
		return IngestResult{}, err
	}

	// Коммит с "анти-призраком"
	if err = tx.Commit(ctx); err != nil {
		// двусмысленность: мог закоммититься
		exists, checkErr := uc.store.ExistsEvent(ctx, e)
		if checkErr == nil && exists {
			// считаем успехом
			if reserved {
				_ = uc.idem.MarkDone(ctx, key)
				committed = true
			}
			slog.Warn("commit error, but event exists; treating as success", "err", err, "event_id", key)
			// продолжим best-effort после коммита — см. ниже
		}

		return IngestResult{}, err
	}

	committed = true
	if reserved {
		_ = uc.idem.MarkDone(ctx, key)
	}

	// ВАЖНО: best-effort вне транзакции
	go func() {
		// отдельный контекст с небольшим тайм-аутом, чтобы не висеть
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		err := uc.store.UpdateUserSignalsBestEffort(ctx, nil, e)
		if err != nil {
			slog.Warn("err store.UpdateUserSignalsBestEffort",
				"event_id", e.EventID,
				"err", err)
		}
	}()

	return IngestResult{Inserted: true}, nil
}
