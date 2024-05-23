package events

import (
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

type baseEvent struct {
	SrcModule  stdtypes.ModuleID
	DestModule stdtypes.ModuleID
	Metadata   map[string]interface{}
}

func newBaseEvent(destModule stdtypes.ModuleID) baseEvent {
	return baseEvent{
		DestModule: destModule,
		Metadata:   make(map[string]interface{}),
	}
}

func (e *baseEvent) Src() stdtypes.ModuleID {
	return e.SrcModule
}

func (e *baseEvent) Dest() stdtypes.ModuleID {
	return e.DestModule
}

func (e *baseEvent) GetMetadata(key string) (interface{}, error) {
	val, ok := e.Metadata[key]
	if !ok {
		return nil, es.Errorf("no metadata for key %s", key)
	}

	return val, nil
}
