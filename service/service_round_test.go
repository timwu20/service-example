package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRoundServiceHappyPath(t *testing.T) {
	sr := NewRoundService(
		func() (serviceA Service, serviceB Service) {
			return NewDelayedStopService(1 * time.Second),
				NewDelayedStopService(1 * time.Second)
		},
	)
	errChan, err := sr.Start()
	if err != nil {
		t.Errorf("%v", err)
	}

	done := make(chan interface{})
	go func() {
		defer close(done)
		time.Sleep(5 * time.Second)
		err := sr.Stop()
		if err != nil {
			t.Errorf("ServiceRound.Stop() error: %v", err)
		}
	}()

	for err := range errChan {
		if err != nil {
			t.Errorf("ServiceRound error: %v", err)
		}
	}
	<-done
}

func TestRoundServiceStartError(t *testing.T) {
	sr := NewRoundService(
		func() (serviceA Service, serviceB Service) {
			return NewStartErrorService(1 * time.Second),
				NewDelayedStopService(1 * time.Second)
		},
	)
	_, err := sr.Start()
	assert.ErrorIs(t, err, ErrStartErrorService)

	sr = NewRoundService(
		func() (serviceA Service, serviceB Service) {
			return NewDelayedStopService(1 * time.Second),
				NewStartErrorService(1 * time.Second)
		},
	)
	_, err = sr.Start()
	assert.ErrorIs(t, err, ErrStartErrorService)
}
