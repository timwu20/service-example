package service

import "fmt"

// Service is a generalized interface for a long running asynchronous service
type Service interface {
	// Start returns an error channel that is only written to if err == nil.  It is implied that
	// the service is ready if the returned err == nil.
	Start() (errChan chan error, err error)
	// Stop will shutdown the service.  The errChan returned in Start() is expected to be closed.
	// Packages that implement the Service interface should implement a timeout to ensure Stop()
	// does not block indefinitely.
	Stop() (err error)
}

type BlockingService interface {
	// Start is a blocking function that will return an error in a fatal error, or nil on
	// successful completion
	Start() (err error)
	// Stop will shutdown the BlockingService.  Packages that implement the SubService interface should implement a timeout to ensure Stop()
	// does not block indefinitely. Will unblock the caller of Start()
	Stop() (err error)
}

var ErrBaseServiceAlreadyStopped = fmt.Errorf("already stopped")

type BaseService struct {
	ErrChan chan error
}

func NewBaseService() *BaseService {
	return &BaseService{
		ErrChan: make(chan error),
	}
}

func (bs *BaseService) Start() (errChan chan error, err error) {
	return bs.ErrChan, nil
}

func (bs *BaseService) Stop() error {
	if bs.ErrChan == nil {
		return ErrBaseServiceAlreadyStopped
	}
	close(bs.ErrChan)
	bs.ErrChan = nil
	return nil
}
