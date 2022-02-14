package service

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSleepPrintService(t *testing.T) {
	sps := NewSleepPrintService(0, time.Millisecond)
	errChan, err := sps.Start()
	if err != nil {
		t.Errorf("%v", err)
	}
	done := make(chan interface{})
	go func() {
		for err := range errChan {
			log.Println(err)
		}
		close(done)
	}()

	time.Sleep(2 * time.Millisecond)
	sps.Stop()
	<-done
}

func TestFanOutServiceService(t *testing.T) {
	sps := NewFanOutService(
		5*time.Millisecond,
		NewSleepPrintService(1, time.Millisecond),
		NewSleepPrintService(2, 500000*time.Nanosecond),
	)
	errChan, err := sps.Start()
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	done := make(chan interface{})
	go func() {
		for err := range errChan {
			log.Println(err)
		}
		close(done)
	}()

	time.Sleep(2 * time.Millisecond)
	sps.Stop()
	<-done
}

func TestFanOutServiceStartError(t *testing.T) {
	sps := NewFanOutService(
		5*time.Millisecond,
		NewSleepPrintService(1, time.Millisecond),
		NewStartErrorService(5*time.Millisecond),
	)
	_, err := sps.Start()
	if err != nil {
		log.Printf("%v\n", err)
	} else {
		t.Errorf("did not receive error")
	}
}

func TestFanOutServiceDelayedError(t *testing.T) {
	sps := NewFanOutService(
		5*time.Millisecond,
		NewSleepPrintService(1, time.Millisecond),
		NewDelayedErrorService(5*time.Millisecond),
	)
	errChan, err := sps.Start()
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	var errCount int
	done := make(chan interface{})
	go func() {
		for err := range errChan {
			errCount++
			assert.Errorf(t, err, fmt.Sprintf("service 1 error: %s", ErrDelayedErrorServiceCritical))

			// SleepPrintService is still running, so it should stop that without error
			// and errChan should close
			err = sps.Stop()
			assert.NoError(t, err)
		}
		close(done)
	}()

	<-done
	assert.Equal(t, 1, errCount)
}
