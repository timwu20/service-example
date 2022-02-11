package service

import (
	"log"
	"sync"
	"time"
)

// SleepPrintService sleeps a configurable amount of time and prints a run count
type SleepPrintService struct {
	index         int
	errChan       chan error
	stopChan      chan interface{}
	wg            sync.WaitGroup
	sleepDuration time.Duration
	runCount      int
}

func NewSleepPrintService(index int, duration time.Duration) *SleepPrintService {
	return &SleepPrintService{
		errChan:       make(chan error),
		stopChan:      make(chan interface{}),
		sleepDuration: duration,
		index:         index,
	}
}

func (sps *SleepPrintService) Start() (errChan chan error, err error) {
	go sps.run()
	return sps.errChan, nil
}

func (sps *SleepPrintService) run() {
	sps.wg.Add(1)
	defer sps.wg.Done()
	ticker := time.NewTicker(sps.sleepDuration)
	for {
		select {
		case <-sps.stopChan:
			return
		case <-ticker.C:
			sps.runCount++
			log.Printf("%d runCount: %d\n", sps.index, sps.runCount)
		}
	}
}

func (sps *SleepPrintService) Stop() (err error) {
	close(sps.stopChan)
	// TODO: implement timeout on wg.wait
	sps.wg.Wait()
	// close the errorChan to unblock any listeners on the errChan
	close(sps.errChan)
	log.Printf("SleepPrintService stopped")
	return
}
