package service

import (
	"fmt"
	"log"
	"time"
)

var ErrDelayedErrorServiceCritical = fmt.Errorf("DelayedErrorService critical error, shutting down")

// StartErrorService will error out on Start after a given duration
type DelayedErrorService struct {
	sleepDuration time.Duration
	errChan       chan error
}

func NewDelayedErrorService(duration time.Duration) *DelayedErrorService {
	return &DelayedErrorService{
		sleepDuration: duration,
		errChan:       make(chan error),
	}
}

func (des *DelayedErrorService) Start() (errChan chan error, err error) {
	go func() {
		time.Sleep(des.sleepDuration)
		errChan <- ErrDelayedErrorServiceCritical
		des.Stop()
	}()
	return des.errChan, nil
}

func (des *DelayedErrorService) Stop() (err error) {
	if des.errChan == nil {
		return fmt.Errorf("DelayedErrorService already stopped")
	}
	log.Println("DelayedErrorService stopped")
	close(des.errChan)
	des.errChan = nil
	return nil
}
