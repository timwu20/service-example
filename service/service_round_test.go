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
		time.Sleep(3 * time.Second)
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

func TestRoundServiceMultipleClose(t *testing.T) {
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
		time.Sleep(3 * time.Second)
		err := sr.Stop()
		if err != nil {
			t.Errorf("ServiceRound.Stop() error: %v", err)
		}
		err = sr.Stop()
		if err != nil {
			assert.ErrorIs(t, err, ErrRoundServiceAlreadyStopped)
		} else {
			t.Errorf("expected error")
		}
	}()

	for err := range errChan {
		if err != nil {
			t.Errorf("ServiceRound error: %v", err)
		}
	}
	<-done
}

func TestRoundServiceDelayedError(t *testing.T) {
	sr := NewRoundService(
		func() (serviceA Service, serviceB Service) {
			return NewDelayedStopService(2 * time.Second),
				NewDelayedErrorService(1 * time.Second)
		},
	)

	errChan, err := sr.Start()
	if err != nil {
		t.Errorf("%v", err)
	}

	for err := range errChan {
		assert.ErrorIs(t, err, ErrDelayedErrorServiceCritical)
	}

	sr = NewRoundService(
		func() (serviceA Service, serviceB Service) {
			return NewDelayedErrorService(1 * time.Second),
				NewDelayedStopService(2 * time.Second)
		},
	)
	errChan, err = sr.Start()
	if err != nil {
		t.Errorf("%v", err)
	}

	for err := range errChan {
		assert.ErrorIs(t, err, ErrDelayedErrorServiceCritical)
	}
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

func TestRoundServiceNewRoundError(t *testing.T) {
	newRoundCallCount := 0
	sr := NewRoundService(
		func() (serviceA Service, serviceB Service) {
			newRoundCallCount++
			switch newRoundCallCount {
			case 1:
				return NewDelayedStopService(1 * time.Second),
					NewDelayedStopService(1 * time.Second)
			default:
				return NewDelayedStopService(1 * time.Second),
					NewStartErrorService(100 * time.Millisecond)
			}
		},
	)
	errChan, err := sr.Start()
	if err != nil {
		t.Errorf("%v", err)
	}

	for err := range errChan {
		assert.ErrorIs(t, err, ErrStartErrorService)
	}
}
