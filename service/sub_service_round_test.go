package service

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBlockingServiceRoundHappyPath(t *testing.T) {
	sr := NewBlockingServiceRound(
		func() (serviceA BlockingService, serviceB BlockingService) {
			return NewBlockingDelayedStopService(1 * time.Second),
				NewBlockingDelayedStopService(1 * time.Second)
		},
	)
	errChan, err := sr.Start()
	if err != nil {
		t.Errorf("%v", err)
	}

	done := make(chan interface{})
	go func() {
		time.Sleep(5 * time.Second)
		err := sr.Stop()
		if err != nil {
			t.Errorf("yo")
		}
		close(done)
	}()

	for err := range errChan {
		if err != nil {
			t.Errorf("ServiceRound error: %v", err)
		}
	}
	<-done
}

func TestBlockingServiceRoundStopOnError(t *testing.T) {
	sr := NewBlockingServiceRound(
		func() (serviceA BlockingService, serviceB BlockingService) {
			return NewBlockingDelayedStopService(2 * time.Second),
				NewBlockingDelayedErrorService(1 * time.Second)
		},
	)
	errChan, err := sr.Start()
	if err != nil {
		t.Errorf("%v", err)
	}

	for err := range errChan {
		assert.ErrorIs(t, err, ErrDelayedErrorServiceCritical)
	}
}

func TestBlockingServiceRoundStopBeforeNextRound(t *testing.T) {
	sr := NewBlockingServiceRound(
		func() (serviceA BlockingService, serviceB BlockingService) {
			return NewBlockingDelayedStopService(1 * time.Second),
				NewBlockingDelayedStopService(1 * time.Second)
		},
	)

	wg := sync.WaitGroup{}

	wg.Add(1)
	sr.roundDone = func() {
		go func() {
			defer wg.Done()
			err := sr.Stop()
			if err != nil {
				t.Errorf("%v", err)
			}
		}()
		time.Sleep(100 * time.Millisecond)
	}

	errChan, err := sr.Start()
	if err != nil {
		t.Errorf("%v", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range errChan {
			if err != nil {
				t.Errorf("ServiceRound error: %v", err)
			}
		}
	}()

	wg.Wait()
}
