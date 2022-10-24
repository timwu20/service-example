package service

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

var ErrRoundServiceAlreadyStopped = ErrBaseServiceAlreadyStopped

type RoundService struct {
	serviceA    Service
	serviceB    Service
	newServices func() (serviceA Service, serviceB Service)
	stopChan    chan interface{}
	wg          sync.WaitGroup
	closed      bool
}

func NewRoundService(newServices func() (serviceA Service, serviceB Service)) (sr *RoundService) {
	return &RoundService{
		newServices: newServices,
		stopChan:    make(chan interface{}),
	}
}

func (rs *RoundService) newRound(errChan chan error) (errChanA, errChanB chan error, err error) {
	serviceA, serviceB := rs.newServices()

	errChanA, err = serviceA.Start()
	if err != nil {
		return
	}

	errChanB, err = serviceB.Start()
	if err != nil {
		stopErr := serviceA.Stop()
		if stopErr != nil {
			log.Printf("error stopping serviceA: %v", stopErr)
		}
		return
	}

	rs.serviceA = serviceA
	rs.serviceB = serviceB
	return
}

func (rs *RoundService) handleErrChans(errChan chan error, errChanA chan error, errChanB chan error) {
	for {
		select {
		case err, ok := <-errChanA:
			if !ok {
				errChanA = nil
			}
			if err != nil {
				errChan <- fmt.Errorf("error from serviceA: %w, stopping RoundService", err)
				go func() {
					err := rs.Stop()
					if err != nil {
						log.Printf("error stopping RoundService: %v\n", err)
					}
				}()
			}
		case err, ok := <-errChanB:
			if !ok {
				errChanB = nil
			}
			if err != nil {
				errChan <- fmt.Errorf("error from serviceB: %w, stopping RoundService", err)
				go func() {
					err := rs.Stop()
					if err != nil {
						log.Printf("error stopping RoundService: %v\n", err)
					}
				}()
			}
		}
		if errChanA == nil && errChanB == nil {
			return
		}
	}
}

func (rs *RoundService) run(errChan chan error, errChanA chan error, errChanB chan error) {
	defer rs.wg.Done()
	defer close(errChan)
	// 1st round
	rs.handleErrChans(errChan, errChanA, errChanB)
	// subsequent rounds
	for {
		select {
		case <-rs.stopChan:
			return
		default:
			errChanA, errChanB, err := rs.newRound(errChan)
			if err != nil {
				errChan <- fmt.Errorf("error calling newRound: %w, stopping RoundService", err)
				go func() {
					err := rs.Stop()
					if err != nil {
						log.Printf("error stopping RoundService: %v\n", err)
					}
				}()
				return
			}
			rs.handleErrChans(errChan, errChanA, errChanB)
		}

	}
}

func (rs *RoundService) Start() (errChan chan error, err error) {
	errChanA, errChanB, err := rs.newRound(errChan)
	if err != nil {
		return
	}

	rs.wg.Add(1)
	errChan = make(chan error)
	go rs.run(errChan, errChanA, errChanB)
	return
}

func (rs *RoundService) Stop() (err error) {
	if rs.closed {
		return ErrBaseServiceAlreadyStopped
	}
	rs.closed = true
	close(rs.stopChan)

	wg := sync.WaitGroup{}
	wg.Add(2)

	errs := [2]error{}
	go func() {
		defer wg.Done()
		errs[0] = rs.serviceA.Stop()
	}()
	go func() {
		defer wg.Done()
		errs[1] = rs.serviceB.Stop()
	}()

	wg.Wait()
	rs.wg.Wait()

	if errs[0] != nil && !errors.Is(errs[0], ErrBaseServiceAlreadyStopped) {
		err = errs[0]
	} else if errs[1] != nil && !errors.Is(errs[1], ErrBaseServiceAlreadyStopped) {
		err = errs[1]
	}
	return
}
