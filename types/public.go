package types

// Common/shared types go here as the SDK grows.
// Keep this file minimal for v0.1.

// Side represents the order side (buy or sell).
type Side int

const (
	SideBuy     Side = 0
	SideSell    Side = 1
	SideUnknown Side = 255
)

// Interval represents time intervals for price history.
type Interval string

const (
	Interval1m  Interval = "1m"
	Interval1h  Interval = "1h"
	Interval6h  Interval = "6h"
	Interval1d  Interval = "1d"
	Interval1w  Interval = "1w"
	IntervalMax Interval = "max"
)
