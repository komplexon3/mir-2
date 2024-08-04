package centraladversay

import (
	"context"
	"fmt"
	"reflect"

	"github.com/filecoin-project/mir/pkg/idledetection"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
	es "github.com/go-errors/errors"
)

var ErrorShutdown = fmt.Errorf("Shutdown")

type IdleNodesMonitor struct {
	idleNotificationC chan chan struct{}
	doneC             chan struct{}
	isDoneC           chan struct{}
	idleDetectionCs   []chan idledetection.IdleNotification
}

func NewIdleNodesMonitor(idleDetectionCs []chan idledetection.IdleNotification) *IdleNodesMonitor {
	return &IdleNodesMonitor{
		idleDetectionCs:   idleDetectionCs,
		idleNotificationC: make(chan chan struct{}),
		doneC:             make(chan struct{}),
		isDoneC:           make(chan struct{}),
	}
}

func (m *IdleNodesMonitor) Stop() {
	close(m.doneC)
	<-m.isDoneC
}

func (m *IdleNodesMonitor) IdleNotificationC() chan chan struct{} {
	return m.idleNotificationC
}

func (m *IdleNodesMonitor) Run(ctx context.Context) error {
	defer close(m.isDoneC)
	var globalIdleNotification chan struct{}

	continueCs := make([]chan idledetection.NoLongerIdleNotification, len(m.idleDetectionCs))
	isIdleStates := make([]bool, len(m.idleDetectionCs))

	var err error
	var ackC chan struct{}
	for err == nil {
		selectCases := make([]reflect.SelectCase, 0, 2*len(m.idleDetectionCs)+2)
		selectReactions := make([]func(receivedVal reflect.Value, ok bool), 0, 2*len(m.idleDetectionCs)+2)

		selectCases = append(selectCases, reflect.SelectCase{
			Chan: reflect.ValueOf(ctx.Done()),
			Dir:  reflect.SelectRecv,
		})
		selectReactions = append(selectReactions, func(_ reflect.Value, _ bool) {
			err = ErrorShutdown
		})

		selectCases = append(selectCases, reflect.SelectCase{
			Chan: reflect.ValueOf(m.doneC),
			Dir:  reflect.SelectRecv,
		})
		selectReactions = append(selectReactions, func(_ reflect.Value, _ bool) {
			err = ErrorShutdown
		})

		for i, continueC := range continueCs {
			selectCases = append(selectCases, reflect.SelectCase{
				Chan: reflect.ValueOf(continueC),
				Dir:  reflect.SelectRecv,
			})
			selectReactions = append(selectReactions, func(_noLongerIdleNotification reflect.Value, ok bool) {
				if !ok {
					return
				}
				noLongerIdleNotification := _noLongerIdleNotification.Interface().(idledetection.NoLongerIdleNotification)
				ackC = noLongerIdleNotification.Ack
				if !isIdleStates[i] {
					err = es.Errorf("node indicating not idle but is already registered as such (node index %d)", i)
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
			selectReactions = append(selectReactions, func(_idleNotification reflect.Value, ok bool) {
				if !ok {
					return
				}
				idleNotification := _idleNotification.Interface().(idledetection.IdleNotification)
				ackC = idleNotification.Ack
				if isIdleStates[i] {
					err = es.Errorf("node indicating idle but is already registered as such (node index %d)", i)
				}

				isIdleStates[i] = true
				continueCs[i] = idleNotification.NoLongerIdleC

				if !sliceutil.Contains(isIdleStates, false) {
					globalIdleNotification = make(chan struct{})
					select {
					case m.idleNotificationC <- globalIdleNotification:
					case <-ctx.Done():
					case <-m.doneC:
					}
				}
			})
		}

		ind, val, ok := reflect.Select(selectCases)
		// channel close
		selectReactions[ind](val, ok)
		if ackC != nil {
			close(ackC)
			ackC = nil
		}
	}

	return err
}
