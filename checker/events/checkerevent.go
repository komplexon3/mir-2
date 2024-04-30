package events

import (
	"encoding/json"
	"fmt"

	"github.com/filecoin-project/mir/stdtypes"
)

type SuccessEvent struct {
	CheckerEvent
}

func NewSuccessEvent() *SuccessEvent {
	return &SuccessEvent{
		NewCheckerEvent(),
	}
}

type FailureEvent struct {
	CheckerEvent
}

func NewFailureEvent() *FailureEvent {
	return &FailureEvent{
		NewCheckerEvent(),
	}
}

type FinalEvent struct {
	CheckerEvent
}

func NewFinalEvent() *FinalEvent {
	return &FinalEvent{
		NewCheckerEvent(),
	}
}

type CheckerEvent struct {
	metadata map[string]interface{}
}

func NewCheckerEvent() CheckerEvent {
	return CheckerEvent{
		metadata: make(map[string]interface{}),
	}
}

func (e *CheckerEvent) Src() stdtypes.ModuleID {
	panic("no source for checker events. only to be used by checker framework")
}

func (e *CheckerEvent) Dest() stdtypes.ModuleID {
	panic("no source for checker events. only to be used by checker framework")
}

func (e *CheckerEvent) ToBytes() ([]byte, error) {
	// really sloppy...
	panic("ToBytes not implemented for checker events. only to be used by checker framework")
	return json.Marshal(e)
}

func (e *CheckerEvent) ToString() string {
	panic("ToString not implemented for checker events. only to be used by checker framework")
	data, err := e.ToBytes()
	if err != nil {
		return fmt.Sprintf("unmarshalableEvent(%+v)", e)
	}

	return string(data)
}

func (e *CheckerEvent) GetMetadata(key string) (interface{}, error) {
	panic("no metadata for checker events. only to be used by checker framework")
}

func (e *CheckerEvent) NewSrc(newSrc stdtypes.ModuleID) stdtypes.Event {
	panic("no src for checker events. only to be used by checker framework")
}

func (e *CheckerEvent) NewDest(newDest stdtypes.ModuleID) stdtypes.Event {
	panic("no dest for checker events. only to be used by checker framework")
}

func (e *CheckerEvent) SetMetadata(key string, value interface{}) (stdtypes.Event, error) {
	panic("no metadata for checker events. only to be used by checker framework")
}
