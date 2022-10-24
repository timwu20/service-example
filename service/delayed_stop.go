package service

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type DelayedStopService struct {
	base          BaseService
	sleepDuration time.Duration
}

func NewDelayedStopService(duration time.Duration) *DelayedStopService {
	return &DelayedStopService{
		base:          *NewBaseService(),
		sleepDuration: duration,
	}
}

func (des *DelayedStopService) Start() (errChan chan error, err error) {
	go func() {
		time.Sleep(des.sleepDuration)
		err := des.Stop()
		if err != nil {
			log.Printf("%v", err)
		}
	}()
	return des.base.ErrChan, nil
}

func (des *DelayedStopService) Stop() (err error) {
	return des.base.Stop()
}

type BlockingDelayedStopService struct {
	sleepDuration time.Duration
	stopChan      chan interface{}
	wg            sync.WaitGroup
}

func NewBlockingDelayedStopService(d time.Duration) *BlockingDelayedStopService {
	return &BlockingDelayedStopService{
		sleepDuration: d,
		stopChan:      make(chan interface{}),
	}
}
func (bdss *BlockingDelayedStopService) Start() (err error) {
	bdss.wg.Add(1)
	defer bdss.wg.Done()
	timer := time.NewTimer(bdss.sleepDuration)
	select {
	case <-timer.C:
	case <-bdss.stopChan:
		timer.Stop()
		return fmt.Errorf("early exit")
	}
	return nil
}

func (bdss *BlockingDelayedStopService) Stop() (err error) {
	if bdss.stopChan == nil {
		return ErrBaseServiceAlreadyStopped
	}
	close(bdss.stopChan)
	bdss.stopChan = nil
	bdss.wg.Wait()
	return
}
