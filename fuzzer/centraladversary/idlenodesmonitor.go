package centraladversay

import (
	"context"
	"fmt"
	"sync"

	es "github.com/go-errors/errors"
)

type IdleNodesMonitor struct {
	idleNotificationC chan chan struct{}
	cancel            context.CancelFunc
	idleDetectionCs   []chan chan struct{}
}

func NewIdleNodesMonitor(idleDetectionCs []chan chan struct{}) *IdleNodesMonitor {
	return &IdleNodesMonitor{
		idleDetectionCs:   idleDetectionCs,
		idleNotificationC: make(chan chan struct{}),
		cancel:            nil,
	}
}

func (m *IdleNodesMonitor) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
}

func (m *IdleNodesMonitor) IdleNotificationC() chan chan struct{} {
	return m.idleNotificationC
}

func (m *IdleNodesMonitor) Run(ctx context.Context) error {
	if m.cancel != nil {
		return es.Errorf("Idle Nodes Monitor is already running or was running")
	}

	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	defer cancel()

	activeC := make(chan struct{})
	idleC := make(chan struct{})
	var noLongerInactiveC chan struct{}

	wg.Add(len(m.idleDetectionCs))
	for _, idleDetectionC := range m.idleDetectionCs {
		go func() {
			defer wg.Done()
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

	var err error
	activeCount := len(m.idleDetectionCs)
ActiveCountLoop:
	for {
		select {
		case <-ctx.Done():
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
				m.idleNotificationC <- noLongerInactiveC
			} else if activeCount < 0 {
				err = es.Errorf("number of active nodes is negative")
				break ActiveCountLoop
			}
		}
	}

	wg.Wait()

	close(activeC)
	close(idleC)
	close(m.idleNotificationC)

	fmt.Println("IdleNodesMonitor done")

	return err
}
