package main

import (
	"context"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func noDelays() func() {
	origRetrySleep := retrySleep
	retrySleep = 0
	return func() {
		retrySleep = origRetrySleep
	}
}

type retryable struct {
	open        bool
	hasClosed   bool
	startedChan chan struct{}
	stopChan    chan error
}

func (r *retryable) Open() error {
	r.open = true
	return nil
}

func (r *retryable) Close() error {
	r.open = false
	r.hasClosed = true
	return nil
}

func (r *retryable) Start(ctx context.Context) error {
	r.startedChan <- struct{}{}
	select {
	case <-ctx.Done():
		r.open = false
		return ctx.Err()
	case err := <-r.stopChan:
		return err
	}
}

func (r *retryable) Name() string {
	return "retryable-test"
}

func TestRetry(t *testing.T) {
	defer noDelays()()
	r := retryable{
		startedChan: make(chan struct{}),
		stopChan:    make(chan error),
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		_ = retry(ctx, &r)
		wg.Done()
	}()
	// wait for start to be called
	<-r.startedChan
	assert.True(t, r.open)

	// trigger start to exit with no error
	r.stopChan <- nil
	<-r.startedChan
	assert.True(t, r.open)

	// emulate an error being returned from start
	r.stopChan <- errors.New("fake error")
	<-r.startedChan
	// check that it was closed and re-opened
	assert.True(t, r.hasClosed)
	assert.True(t, r.open)

	cancel()
	wg.Wait()
}
