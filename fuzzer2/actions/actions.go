package actions

import "github.com/filecoin-project/mir/stdtypes"

type Actions interface {
	SelectAction() Action
}

// Returns: events to forward/inject, whether or not this makes this node byzantine, what to log this action as, error
type Action func(event stdtypes.Event, sourceNode stdtypes.NodeID, byzantineNodes []stdtypes.NodeID) (string, map[stdtypes.NodeID]*stdtypes.EventList, error)
