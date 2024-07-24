package stdevents

import (
	t "github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

type mirEvent struct {
	SrcModule  t.ModuleID
	DestModule t.ModuleID
	Metadata   map[string]interface{}
}

func newMirEvent(dest t.ModuleID) mirEvent {
	return mirEvent{
		DestModule: dest,
		Metadata:   make(map[string]interface{}),
	}
}

func (e *mirEvent) Src() t.ModuleID {
	return e.SrcModule
}

func (e *mirEvent) Dest() t.ModuleID {
	return e.DestModule
}

func (e *mirEvent) GetMetadata(key string) (interface{}, error) {
	val, ok := e.Metadata[key]
	if !ok {
		return nil, es.Errorf("no metadata for key %s", key)
	}

	return val, nil
}
