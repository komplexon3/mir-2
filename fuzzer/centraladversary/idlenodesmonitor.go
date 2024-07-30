package centraladversay

import (
	"context"
	"sync"

	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

type IdleNodesMonitor struct {
	idleNotificationC chan chan struct{}
	cancel            context.CancelFunc
	idleDetectionCs   map[stdtypes.NodeID]chan chan struct{}
}

func NewIdleNodesMonitor(idleDetectionCs map[stdtypes.NodeID]chan chan struct{}) *IdleNodesMonitor {
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

	activeC := make(chan stdtypes.NodeID)
	idleC := make(chan stdtypes.NodeID)
	var noLongerInactiveC chan struct{}

	wg.Add(len(m.idleDetectionCs))
	for nodeID, idleDetectionC := range m.idleDetectionCs {
		go func() {
			defer wg.Done()
			var continueC chan struct{}
			for {
				select {
				case <-ctx.Done():
					return
				case <-continueC:
					activeC <- nodeID
					continueC = nil
				case cc := <-idleDetectionC:
					idleC <- nodeID
					continueC = cc
				}
			}
		}()
	}

	var err error
	isIdleStates := make(map[stdtypes.NodeID]bool)
	for nodeID := range m.idleDetectionCs {
		isIdleStates[nodeID] = false
	}
ActiveCountLoop:
	for {
		select {
		case <-ctx.Done():
			break ActiveCountLoop
		case nodeID := <-activeC:
			if noLongerInactiveC != nil {
				close(noLongerInactiveC)
				noLongerInactiveC = nil
			}
			if !isIdleStates[nodeID] {
				err = es.Errorf("node %s indicating not idle but is already registered as such", nodeID)
				break ActiveCountLoop
			}
			isIdleStates[nodeID] = false
		case nodeID := <-idleC:
			if isIdleStates[nodeID] {
				err = es.Errorf("node %s indicating idle but is already registered as such", nodeID)
				break ActiveCountLoop
			}

			isIdleStates[nodeID] = true

			if !sliceutil.Contains(maputil.GetValues(isIdleStates), false) {
				noLongerInactiveC = make(chan struct{})
				m.idleNotificationC <- noLongerInactiveC
			}
		}
	}

	wg.Wait()

	close(activeC)
	close(idleC)
	close(m.idleNotificationC)

	return err
}
