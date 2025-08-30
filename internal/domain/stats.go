package domain

import "time"

// StatsTotals агрегированные метрики за период
type StatsTotals struct {
	Views       int64
	Completes   int64
	Likes       int64
	Clicks      int64
	Impressions int64
	DwellMsSum  int64
}

// StatsKPI рассчитываемые KPI
type StatsKPI struct {
	CTR       float64 // clicks / impressions
	AvgViewMs float64 // dwell_ms_sum / views
}

// VideoWithStats запись топ-видео
type VideoWithStats struct {
	Video       Video
	Views       int64
	Completes   int64
	Likes       int64
	Clicks      int64
	Impressions int64
	Score       float64
}

// StatsOverview ответ ручки
type StatsOverview struct {
	From      time.Time
	To        time.Time
	Totals    StatsTotals
	KPI       StatsKPI
	TopVideos []VideoWithStats
}
