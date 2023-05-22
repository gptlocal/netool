package backoff

import (
	"github.com/gptlocal/netool/w/grpc/internal/grpcrand"
	"time"
)

// Config defines the configuration options for backoff.
type Config struct {
	// BaseDelay is the amount of time to backoff after the first failure.
	BaseDelay time.Duration
	// Multiplier is the factor with which to multiply backoffs after a failed retry. Should ideally be greater than 1.
	Multiplier float64
	// Jitter is the factor with which backoffs are randomized.
	Jitter float64
	// MaxDelay is the upper bound of backoff delay.
	MaxDelay time.Duration
}

// DefaultConfig is a backoff configuration with the default values specfied
// at https://github.com/grpc/grpc/blob/master/doc/connection-backoff.md.
//
// This should be useful for callers who want to configure backoff with
// non-default values only for a subset of the options.
var DefaultConfig = Config{
	BaseDelay:  1.0 * time.Second,
	Multiplier: 1.6,
	Jitter:     0.2,
	MaxDelay:   120 * time.Second,
}

// Strategy defines the methodology for backing off after a grpc connection failure.
type Strategy interface {
	// Backoff returns the amount of time to wait before the next retry given the number of consecutive failures.
	Backoff(retries int) time.Duration
}

// DefaultExponential is an exponential backoff implementation using the
// default values for all the configurable knobs defined in
// https://github.com/grpc/grpc/blob/master/doc/connection-backoff.md.
var DefaultExponential = Exponential{Config: DefaultConfig}

// Exponential implements exponential backoff algorithm as defined in
// https://github.com/grpc/grpc/blob/master/doc/connection-backoff.md.
type Exponential struct {
	// Config contains all options to configure the backoff algorithm.
	Config Config
}

// Backoff returns the amount of time to wait before the next retry given the number of retries.
func (bc Exponential) Backoff(retries int) time.Duration {
	if retries == 0 {
		return bc.Config.BaseDelay
	}
	backoff, max := float64(bc.Config.BaseDelay), float64(bc.Config.MaxDelay)
	for backoff < max && retries > 0 {
		backoff *= bc.Config.Multiplier
		retries--
	}
	if backoff > max {
		backoff = max
	}
	// Randomize backoff delays so that if a cluster of requests start at the same time, they won't operate in lockstep.
	backoff *= 1 + bc.Config.Jitter*(grpcrand.Float64()*2-1)
	if backoff < 0 {
		return 0
	}
	return time.Duration(backoff)
}
