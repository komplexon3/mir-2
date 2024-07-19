package actions

import (
	"github.com/filecoin-project/mir/stdtypes"
)

type Actions interface {
	SelectAction() Action
}

type DelayedEvents struct {
	Events *stdtypes.EventList
	NodeID stdtypes.NodeID
}

// Returns: events to forward/inject, whether or not this makes this node byzantine, what to log this action as, error
type Action func(event stdtypes.Event, sourceNode stdtypes.NodeID, byzantineNodes []stdtypes.NodeID) (string, map[stdtypes.NodeID]*stdtypes.EventList, []DelayedEvents, error)
