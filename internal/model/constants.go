package model

import "time"

const (
	DefaultLeaseTTL        = 30 * time.Second
	PreheatWindow            = 5 * time.Minute
	IgnitionDelayWindow    = 15 * time.Second
	GlycolSwellSettleWindow  = 45 * time.Second
	HeatbankWarmupWindow = 2 * time.Minute
	FeedwaterRampWindow    = 30 * time.Second
	MaxGlycolLevelPercent    = 95.0
	MinGlycolLevelPercent    = 15.0
	TripGlycolLowPercent     = 10.0
	TripGlycolHighPercent    = 98.0
	NormalSteamPressurePSI = 1800.0
	MaxSteamPressurePSI    = 2000.0
	MinZonelockO2Percent    = 2.5
	MaxZonelockO2Percent    = 6.0
	DefaultJournalCapacity = 512
)
