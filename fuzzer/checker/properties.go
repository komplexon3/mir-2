package checker

import (
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/stdtypes"
)

type (
	Property   = modules.PassiveModule
	Properties = map[string]modules.PassiveModule
)

type property struct {
	runnerModule modules.PassiveModule
	eventChan    chan stdtypes.Event
	isDoneC      chan struct{}
	name         string
	result       CheckerResult
	done         bool
}

func newProperty(name string, module modules.PassiveModule) *property {
	return &property{
		name:         name,
		runnerModule: module,
		eventChan:    make(chan stdtypes.Event),
		isDoneC:      make(chan struct{}, 1), // TODO: why does this channel have a capacity of 1?
		done:         false,
		result:       NOT_READY,
	}
}
