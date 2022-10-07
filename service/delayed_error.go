package service

import (
	"fmt"
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
