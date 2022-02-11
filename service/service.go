package service

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Service interface {
	Start() (errs chan error, err error)
	Stop() (err error)
}

type SleepPrintService struct {
	index         int
	errChan       chan error
	stopChan      chan interface{}
	wg            sync.WaitGroup
	sleepDuration time.Duration
	runCount      int
}

func NewSleepPrintService(index int, duration time.Duration) *SleepPrintService {
	return &SleepPrintService{
		errChan:       make(chan error),
		stopChan:      make(chan interface{}),
		sleepDuration: duration,
		index:         index,
	}
}

func (sps *SleepPrintService) Start() (errChan chan error, err error) {
	go sps.run()
	return sps.errChan, nil
}

func (sps *SleepPrintService) run() {
	sps.wg.Add(1)
	defer sps.wg.Done()
	ticker := time.NewTicker(sps.sleepDuration)
	for {
		select {
		case <-sps.stopChan:
			return
		case <-ticker.C:
			sps.runCount++
			log.Printf("%d runCount: %d\n", sps.index, sps.runCount)
		}
	}
}

func (sps *SleepPrintService) Stop() (err error) {
	close(sps.stopChan)
	// TODO: implement timeout on wg.wait
	sps.wg.Wait()
	// close the errorChan to unblock any listeners on the errChan
	close(sps.errChan)
	return
}

type StartErrorService struct {
	sleepDuration time.Duration
}

func NewStartErrorService(duration time.Duration) *StartErrorService {
	return &StartErrorService{
		sleepDuration: duration,
	}
}

func (ses *StartErrorService) Start() (errChan chan error, err error) {
	time.Sleep(ses.sleepDuration)
	return nil, fmt.Errorf("FATAL: StartErrorService fatal error")
}

func (ses *StartErrorService) Stop() (err error) {
	log.Panicf("huh? shouldn't be called")
	return
}

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
	// errChans := make([]chan error, 0)
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
	// TODO: implement timeout if any service is failing, and shutdown any running services

	if len(startErrs) != 0 {
		err := fos.Stop()
		if err != nil {
			log.Printf("Error stopping services after receiving startErrs: %v", err)
		}
		return nil, startErrs[0]
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

	if len(stopErrs) > 0 {
		return stopErrs[0]
	}
	return
}
