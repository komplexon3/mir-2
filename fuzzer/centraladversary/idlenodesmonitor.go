package centraladversay

import (
	"context"
	"sync"

	es "github.com/go-errors/errors"
)

type IdleNodesMonitor struct {
	idleNotificationC chan chan struct{}
	doneC             chan struct{}
	wg                sync.WaitGroup
}

func NewIdleNodesMonitor() *IdleNodesMonitor {
	return &IdleNodesMonitor{
		idleNotificationC: make(chan chan struct{}),
		doneC:             make(chan struct{}),
		wg:                sync.WaitGroup{},
	}
}

func (inm *IdleNodesMonitor) Stop() {
	close(inm.doneC)
	inm.wg.Wait()
}

func (inm *IdleNodesMonitor) IdleNotificationC() chan chan struct{} {
	return inm.idleNotificationC
}

func (inm *IdleNodesMonitor) Run(ctx context.Context, idleDetectionCs []chan chan struct{}) error {
	ctx, cancel := context.WithCancel(ctx)
	activeC := make(chan struct{})
	idleC := make(chan struct{})
	defer close(activeC)
	defer close(idleC)
	defer close(inm.idleNotificationC)
	var noLongerInactiveC chan struct{}
	defer cancel()
	defer func() {
		for _, nc := range idleDetectionCs {
			close(nc)
		}
	}()

	inm.wg.Add(len(idleDetectionCs))
	for _, idleDetectionC := range idleDetectionCs {
		go func() {
			defer inm.wg.Done()
			var continueC chan struct{}
			for {
				select {
				case <-ctx.Done():
					return
				case <-continueC:
					activeC <- struct{}{}
					continueC = nil
				case cc := <-idleDetectionC:
					idleC <- struct{}{}
					continueC = cc
				}
			}
		}()
	}

	activeCount := len(idleDetectionCs)
ActiveCountLoop:
	for {
		select {
		case <-ctx.Done():
			break ActiveCountLoop
		case <-inm.doneC:
			break ActiveCountLoop
		case <-activeC:
			if noLongerInactiveC != nil {
				close(noLongerInactiveC)
				noLongerInactiveC = nil
			}
			activeCount++
		case <-idleC:
			activeCount--
			if activeCount == 0 {
				noLongerInactiveC = make(chan struct{})
				inm.idleNotificationC <- noLongerInactiveC
			} else if activeCount < 0 {
				return es.Errorf("number of active nodes is negative")
			}
		}
	}

	inm.wg.Wait()
	return nil
}
