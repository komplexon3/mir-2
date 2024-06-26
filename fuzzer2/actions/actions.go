package actions

import "github.com/filecoin-project/mir/stdtypes"

type Actions interface {
	SelectAction() Action
}

type Action func(event stdtypes.Event) (*stdtypes.EventList, error)
