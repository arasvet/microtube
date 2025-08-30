package usecase

import (
	"context"
	"time"

	"github.com/arasvet/microtube/internal/domain"
	"github.com/arasvet/microtube/internal/repo"
)

type StatsUCInterface interface {
	Overview(ctx context.Context, from, to time.Time, top int) (domain.StatsOverview, error)
}

type StatsUC struct {
	store repo.Store
}

func NewStatsUC(store repo.Store) *StatsUC { return &StatsUC{store: store} }

func (uc *StatsUC) Overview(ctx context.Context, from, to time.Time, top int) (domain.StatsOverview, error) {
	fromStr := from.UTC().Format("2006-01-02")
	toStr := to.UTC().Format("2006-01-02")

	totals, err := uc.store.StatsTotals(ctx, fromStr, toStr)
	if err != nil {
		return domain.StatsOverview{}, err
	}
	topVideos, err := uc.store.StatsTopVideos(ctx, fromStr, toStr, top)
	if err != nil {
		return domain.StatsOverview{}, err
	}

	kpi := domain.StatsKPI{}
	if totals.Impressions > 0 {
		kpi.CTR = float64(totals.Clicks) / float64(totals.Impressions)
	}
	if totals.Views > 0 {
		kpi.AvgViewMs = float64(totals.DwellMsSum) / float64(totals.Views)
	}

	return domain.StatsOverview{
		From:      from,
		To:        to,
		Totals:    totals,
		KPI:       kpi,
		TopVideos: topVideos,
	}, nil
}
