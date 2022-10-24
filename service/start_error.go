package service

import (
	"fmt"
	"time"
)

var ErrStartErrorService = fmt.Errorf("FATAL: StartErrorService fatal error")

// StartErrorService will error out on Start after a given duration
type StartErrorService struct {
	BaseService
	sleepDuration time.Duration
}

func NewStartErrorService(duration time.Duration) *StartErrorService {
	return &StartErrorService{
		BaseService:   *NewBaseService(),
		sleepDuration: duration,
	}
}

func (ses *StartErrorService) Start() (errChan chan error, err error) {
	time.Sleep(ses.sleepDuration)
	return nil, ErrStartErrorService
}

func (ses *StartErrorService) Stop() (err error) {
	return ses.BaseService.Stop()
}
