package service

import (
	"log"
	"testing"
	"time"
)

func TestSleepPrintService(t *testing.T) {
	sps := NewSleepPrintService(0, time.Second)
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

	time.Sleep(10 * time.Second)
	sps.Stop()
	<-done
}

func TestFanOutServiceService(t *testing.T) {
	sps := NewFanOutService(
		NewSleepPrintService(1, time.Second),
		NewSleepPrintService(2, 500*time.Millisecond),
	)
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

	time.Sleep(5 * time.Second)
	sps.Stop()
	<-done
}

func TestFanOutServiceStartError(t *testing.T) {
	sps := NewFanOutService(
		NewSleepPrintService(1, time.Second),
		NewStartErrorService(3*time.Second),
	)
	_, err := sps.Start()
	if err != nil {
		log.Printf("%v\n", err)
	} else {
		t.Errorf("did not receive error")
	}

}
