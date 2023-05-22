package health

import (
	"context"
	"errors"
	"google.golang.org/grpc/connectivity"
	"reflect"
	"testing"
	"time"
)

const defaultTestTimeout = 10 * time.Second

func (s) TestClientHealthCheckBackoff(t *testing.T) {
	const maxRetries = 5

	var want []time.Duration
	for i := 0; i < maxRetries; i++ {
		want = append(want, time.Duration(i+1)*time.Second)
	}

	var got []time.Duration
	newStream := func(string) (interface{}, error) {
		if len(got) < maxRetries {
			return nil, errors.New("backoff")
		}
		return nil, nil
	}

	oldBackoffFunc := backoffFunc
	backoffFunc = func(ctx context.Context, retries int) bool {
		got = append(got, time.Duration(retries+1)*time.Second)
		return true
	}
	defer func() { backoffFunc = oldBackoffFunc }()

	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()
	clientHealthCheck(ctx, newStream, func(connectivity.State, error) {}, "test")

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Backoff durations for %v retries are %v. (expected: %v)", maxRetries, got, want)
	}
}
