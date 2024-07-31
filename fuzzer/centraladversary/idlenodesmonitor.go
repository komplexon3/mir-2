package centraladversay

import (
	"context"
	"fmt"
	"reflect"

	idledetection "github.com/filecoin-project/mir/pkg/idleDetection"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
	es "github.com/go-errors/errors"
)

var ErrorShutdown = fmt.Errorf("Shutdown")

type IdleNodesMonitor struct {
	idleNotificationC chan idledetection.IdleNotification
	cancel            context.CancelFunc
	idleDetectionCs   []chan idledetection.IdleNotification
}

func NewIdleNodesMonitor(idleDetectionCs []chan idledetection.IdleNotification) *IdleNodesMonitor {
	return &IdleNodesMonitor{
		idleDetectionCs:   idleDetectionCs,
		idleNotificationC: make(chan idledetection.IdleNotification),
		cancel:            nil,
	}
}

func (m *IdleNodesMonitor) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
}

func (m *IdleNodesMonitor) IdleNotificationC() chan idledetection.IdleNotification {
	return m.idleNotificationC
}

func (m *IdleNodesMonitor) Run(ctx context.Context) error {
	if m.cancel != nil {
		return es.Errorf("Idle Nodes Monitor is already running or was running")
	}

	ctx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	defer cancel()

	continueCs := make([]chan idledetection.NoLongerIdleNotification, len(m.idleDetectionCs))

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
			selectReactions = append(selectReactions, func(_noLongerIdleNotification reflect.Value) {
				noLongerIdleNotification := _noLongerIdleNotification.Interface().(idledetection.NoLongerIdleNotification)

				if !isIdleStates[i] {
					err = es.Errorf("node indicating not idle but is already registered as such")
				}
				isIdleStates[i] = false
				continueCs[i] = nil
				close(noLongerIdleNotification.Ack)
			})
		}

		for i, idleDetectionC := range m.idleDetectionCs {
			selectCases = append(selectCases, reflect.SelectCase{
				Chan: reflect.ValueOf(idleDetectionC),
				Dir:  reflect.SelectRecv,
			})
			selectReactions = append(selectReactions, func(_idleNotification reflect.Value) {
				idleNotification := _idleNotification.Interface().(idledetection.IdleNotification)
				if isIdleStates[i] {
					err = es.Errorf("node indicating idle but is already registered as such")
				}

				isIdleStates[i] = true
				continueCs[i] = idleNotification.NoLongerIdleC

				if !sliceutil.Contains(isIdleStates, false) &&
					len(sliceutil.Filter(continueCs, func(_ int, c chan idledetection.NoLongerIdleNotification) bool { return len(c) > 0 })) == 0 {
					globalIdleNotification := idledetection.NewIdleNotification()
					select {
					case m.idleNotificationC <- globalIdleNotification:
					case <-ctx.Done():
					}
					select {
					case <-globalIdleNotification.Ack:
					case <-ctx.Done():
					}
				}
				close(idleNotification.Ack)
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

	close(m.idleNotificationC)

	if err != nil && err != ErrorShutdown {
		return err
	}

	return nil
}
