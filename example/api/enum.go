package api

type Status int

const (
	StatusUnavailable Status = iota // unavailable
	StatusAvailable                 // available
)
