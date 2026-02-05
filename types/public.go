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
