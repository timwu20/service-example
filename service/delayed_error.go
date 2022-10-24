package service

import (
	"fmt"
	"sync"
	"time"
)

var ErrDelayedErrorServiceCritical = fmt.Errorf("DelayedErrorService critical error, shutting down")

// StartErrorService will error out on Start after a given duration
type DelayedErrorService struct {
	base          BaseService
	sleepDuration time.Duration
}

func NewDelayedErrorService(duration time.Duration) *DelayedErrorService {
	return &DelayedErrorService{
		base:          *NewBaseService(),
		sleepDuration: duration,
	}
}

func (des *DelayedErrorService) Start() (errChan chan error, err error) {
	go func() {
		time.Sleep(des.sleepDuration)
		des.base.ErrChan <- ErrDelayedErrorServiceCritical
		des.Stop()
	}()
	return des.base.ErrChan, nil
}

func (des *DelayedErrorService) Stop() (err error) {
	return des.base.Stop()
}

type BlockingDelayedErrorService struct {
	sleepDuration time.Duration
	stopChan      chan interface{}
	wg            sync.WaitGroup
}

func NewBlockingDelayedErrorService(d time.Duration) *BlockingDelayedErrorService {
	return &BlockingDelayedErrorService{
		sleepDuration: d,
		stopChan:      make(chan interface{}),
	}
}
func (bdss *BlockingDelayedErrorService) Start() (err error) {
	bdss.wg.Add(1)
	defer bdss.wg.Done()
	timer := time.NewTimer(bdss.sleepDuration)
	select {
	case <-timer.C:
		return ErrDelayedErrorServiceCritical
	case <-bdss.stopChan:
		timer.Stop()
		return fmt.Errorf("early exit")
	}
}

func (bdss *BlockingDelayedErrorService) Stop() (err error) {
	if bdss.stopChan == nil {
		return ErrBaseServiceAlreadyStopped
	}
	close(bdss.stopChan)
	bdss.stopChan = nil
	bdss.wg.Wait()
	return
}
