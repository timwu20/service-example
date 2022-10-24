package service

import (
	"errors"
	"sync"
)

var ErrSubServiceRoundAlreadyStopped = ErrBaseServiceAlreadyStopped

type BlockingServiceRound struct {
	serviceA    BlockingService
	serviceB    BlockingService
	newServices func() (serviceA BlockingService, serviceB BlockingService)
	stopChan    chan interface{}
	wg          sync.WaitGroup
	closed      bool
	roundDone   func()
	sync.Mutex
}

func NewBlockingServiceRound(newServices func() (serviceA BlockingService, serviceB BlockingService)) (sr *BlockingServiceRound) {
	return &BlockingServiceRound{
		newServices: newServices,
		stopChan:    make(chan interface{}),
	}
}

func (rs *BlockingServiceRound) Start() (errChan chan error, err error) {
	errChan = make(chan error)

	rs.wg.Add(1)

	go func() {
		defer close(errChan)
		defer rs.wg.Done()

	round:
		for {
			select {
			case <-rs.stopChan:
				break round
			default:
			}

			rs.Lock()
			rs.serviceA, rs.serviceB = rs.newServices()
			rs.Unlock()

			errsA := make(chan error)
			errsB := make(chan error)
			go func() {
				defer close(errsA)
				err := rs.serviceA.Start()
				if err != nil {
					errsA <- err
				}
			}()
			go func() {
				defer close(errsB)
				err := rs.serviceB.Start()
				if err != nil {
					errsB <- err
				}
			}()

		poll:
			for {
				select {
				case <-rs.stopChan:
					break round
				case err, ok := <-errsA:
					if !ok {
						errsA = nil
					}
					if err != nil {
						errChan <- err
						go rs.Stop()
					}
				case err, ok := <-errsB:
					if !ok {
						errsB = nil
					}
					if err != nil {
						errChan <- err
						go rs.Stop()
					}
				default:
					if errsA == nil && errsB == nil {
						if rs.roundDone != nil {
							rs.roundDone()
						}
						break poll
					}
				}
			}
		}
	}()

	return
}

func (rs *BlockingServiceRound) Stop() (err error) {
	if rs.closed {
		return ErrBaseServiceAlreadyStopped
	}
	rs.closed = true
	close(rs.stopChan)

	wg := sync.WaitGroup{}
	wg.Add(2)
	rs.Lock()

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
	rs.Unlock()

	rs.wg.Wait()

	if errs[0] != nil && !errors.Is(errs[0], ErrBaseServiceAlreadyStopped) {
		err = errs[0]
	} else if errs[1] != nil && !errors.Is(errs[1], ErrBaseServiceAlreadyStopped) {
		err = errs[1]
	}
	return
}
