package centraladversay

import (
	"math/rand/v2"

	"github.com/filecoin-project/mir/fuzzer/actions"
)

type delayedEventsManager struct {
	delayedEvents []actions.DelayedEvents
}

func NewDelayedEventsManager() *delayedEventsManager {
	return &delayedEventsManager{
		delayedEvents: make([]actions.DelayedEvents, 0),
	}
}

func (m *delayedEventsManager) Push(delayedEvents ...actions.DelayedEvents) {
	m.delayedEvents = append(m.delayedEvents, delayedEvents...)
}

func (m *delayedEventsManager) PopRandomDelayedEvents() actions.DelayedEvents {
	ind := rand.IntN(len(m.delayedEvents))
	de := m.delayedEvents[ind]
	// scrambeling order, ok bc we are picking random elements anyways
	m.delayedEvents[ind] = m.delayedEvents[len(m.delayedEvents)-1]
	m.delayedEvents = m.delayedEvents[:len(m.delayedEvents)-1]
	return de
}

func (m *delayedEventsManager) Empty() bool {
	return len(m.delayedEvents) == 0
}
