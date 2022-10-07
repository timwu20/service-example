package service

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

type RoundService struct {
	serviceA    Service
	serviceB    Service
	newServices func() (serviceA Service, serviceB Service)
	stopChan    chan interface{}
	wg          sync.WaitGroup
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
				err := rs.Stop()
				if err != nil {
					log.Printf("error stopping RoundService: %v\n", err)
				}
			}
		case err, ok := <-errChanB:
			if !ok {
				errChanB = nil
			}
			if err != nil {
				errChan <- fmt.Errorf("error from serviceB: %w, stopping RoundService", err)
				err := rs.Stop()
				if err != nil {
					log.Printf("error stopping RoundService: %v\n", err)
				}
			}
		}
		if errChanA == nil && errChanB == nil {
			break
		}
	}
}

func (sr *RoundService) Start() (errChan chan error, err error) {
	errChanA, errChanB, err := sr.newRound(errChan)
	if err != nil {
		return
	}

	sr.wg.Add(1)
	errChan = make(chan error)
	go func() {
		defer sr.wg.Done()
		defer close(errChan)
		// 1st round
		sr.handleErrChans(errChan, errChanA, errChanB)
		// subsequent rounds
		for {
			select {
			case <-sr.stopChan:
				return
			default:
				errChanA, errChanB, err := sr.newRound(errChan)
				if err != nil {
					errChan <- fmt.Errorf("error calling newRound: %w, stopping RoundService", err)
					err := sr.Stop()
					if err != nil {
						log.Printf("error stopping RoundService: %v\n", err)
					}
					return
				}
				sr.handleErrChans(errChan, errChanA, errChanB)
			}

		}
	}()
	return
}

func (sr *RoundService) Stop() (err error) {
	close(sr.stopChan)

	wg := sync.WaitGroup{}
	wg.Add(2)

	errs := [2]error{}
	go func() {
		defer wg.Done()
		errs[0] = sr.serviceA.Stop()
	}()
	go func() {
		defer wg.Done()
		errs[1] = sr.serviceB.Stop()
	}()

	wg.Wait()
	sr.wg.Wait()

	if errs[0] != nil && !errors.Is(errs[0], ErrBaseServiceAlreadyStopped) {
		err = errs[0]
	} else if errs[1] != nil && !errors.Is(errs[1], ErrBaseServiceAlreadyStopped) {
		err = errs[1]
	}
	return
}
