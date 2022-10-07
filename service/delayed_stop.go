package service

import (
	"log"
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
