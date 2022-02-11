package service

import (
	"fmt"
	"log"
	"sync"
)

// FanOutService takes an unbounded number of Service implementations and starts
// and stops them in parallel
type FanOutService struct {
	services []Service
	running  []Service
	errChan  chan error
	wg       sync.WaitGroup
}

func NewFanOutService(services ...Service) *FanOutService {
	return &FanOutService{
		services: services,
		errChan:  make(chan error),
	}
}

func (fos *FanOutService) handleServiceErrors(i int, errChan chan error) {
	fos.wg.Add(1)
	defer fos.wg.Done()

	for err := range errChan {
		fos.errChan <- fmt.Errorf("service %d error: %v", i, err)
	}
}

func (fos *FanOutService) Start() (errChan chan error, err error) {
	var startWG sync.WaitGroup
	startErrs := make([]error, 0)
	for i := range fos.services {
		startWG.Add(1)
		go func(i int, s Service) {
			defer startWG.Done()
			errChan, err := s.Start()
			if err != nil {
				startErrs = append(startErrs, err)
				return
			}
			fos.running = append(fos.running, s)
			go fos.handleServiceErrors(i, errChan)
		}(i, fos.services[i])
	}
	startWG.Wait()

	if len(startErrs) != 0 {
		err := fos.Stop()
		if err != nil {
			// TODO: join errors to combine start and stop error
			log.Printf("Error stopping services after receiving startErrs: %v", err)
		}
		return nil, fmt.Errorf("received an error when starting services: %v", startErrs[0])
	}
	return fos.errChan, nil
}

func (fos *FanOutService) Stop() (err error) {
	stopErrs := make([]error, 0)
	var wg sync.WaitGroup
	for i := range fos.running {
		wg.Add(1)
		go func(i int, s Service) {
			defer wg.Done()
			err := s.Stop()
			if err != nil {
				stopErrs = append(stopErrs, err)
			}
		}(i, fos.running[i])
	}
	wg.Wait()
	fos.wg.Wait()
	close(fos.errChan)

	log.Printf("FanOutService stopped")

	if len(stopErrs) > 0 {
		return stopErrs[0]
	}
	return
}
