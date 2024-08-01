package centraladversay

import (
	"context"
	"fmt"
	"reflect"

	"github.com/filecoin-project/mir/pkg/util/sliceutil"
	es "github.com/go-errors/errors"
)

var ErrorShutdown = fmt.Errorf("Shutdown")

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

	ctx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	defer cancel()

	var globalIdleNotification chan struct{}

	continueCs := make([]chan struct{}, len(m.idleDetectionCs))
	isIdleStates := make([]bool, len(m.idleDetectionCs))

	var err error
	for err == nil {
		selectCases := make([]reflect.SelectCase, 0, 2*len(m.idleDetectionCs)+1)
		selectReactions := make([]func(receivedVal reflect.Value), 0, 2*len(m.idleDetectionCs)+1)

		for i, continueC := range continueCs {
			selectCases = append(selectCases, reflect.SelectCase{
				Chan: reflect.ValueOf(continueC),
				Dir:  reflect.SelectRecv,
			})
			selectReactions = append(selectReactions, func(_ reflect.Value) {
				if !isIdleStates[i] {
					err = es.Errorf("node indicating not idle but is already registered as such")
				}
				isIdleStates[i] = false
				continueCs[i] = nil
			})
		}

		for i, idleDetectionC := range m.idleDetectionCs {
			selectCases = append(selectCases, reflect.SelectCase{
				Chan: reflect.ValueOf(idleDetectionC),
				Dir:  reflect.SelectRecv,
			})
			selectReactions = append(selectReactions, func(_continueC reflect.Value) {
				if isIdleStates[i] {
					err = es.Errorf("node indicating idle but is already registered as such")
				}

				isIdleStates[i] = true
				continueCs[i] = _continueC.Interface().(chan struct{})

				if !sliceutil.Contains(isIdleStates, false) {
					globalIdleNotification = make(chan struct{})
					select {
					case m.idleNotificationC <- globalIdleNotification:
					case <-ctx.Done():
					}
				}
			})
		}

		selectCases = append(selectCases, reflect.SelectCase{
			Chan: reflect.ValueOf(ctx.Done()),
			Dir:  reflect.SelectRecv,
		})
		selectReactions = append(selectReactions, func(_ reflect.Value) {
			err = ErrorShutdown
		})

		ind, val, _ := reflect.Select(selectCases)
		selectReactions[ind](val)
	}

	return err
}
