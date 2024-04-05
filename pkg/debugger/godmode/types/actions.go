package types

import (
	"encoding/json"
	"fmt"

	"github.com/filecoin-project/mir/stdtypes"
)

type NodeInterfaceState string

const (
	Open     NodeInterfaceState = "Open"
	Blocking NodeInterfaceState = "Blocking"
	Dropping NodeInterfaceState = "Dropping"
)

type InterfaceChange struct {
	InterfaceState NodeInterfaceState `json:"InterfaceState"`
}

func (*InterfaceChange) Type() GodModeActionTypes {
	return TypeInterfaceChange
}

type DispatchItem struct {
	Node      stdtypes.NodeID
	MessageId string // probably some kind of digest
}
type Dispatch struct {
	Items []DispatchItem `json:"Items"`
}

func (*Dispatch) Type() GodModeActionTypes {
	return TypeDispatch
}

type GodModeActionTypes string

type PlayPauseOption string

const (
	Play  PlayPauseOption = "Play"
	Pause PlayPauseOption = "Pause"
)

type PlayPause struct {
	Value PlayPauseOption `json:"Value"`
}

func (*PlayPause) Type() GodModeActionTypes {
	return TypePlayPause
}

const (
	TypeInterfaceChange GodModeActionTypes = "InterfaceChange"
	TypeDispatch        GodModeActionTypes = "Dispatch"
	TypePlayPause       GodModeActionTypes = "PlayPause"
)

type GodModeAction interface {
	Type() GodModeActionTypes
}

type GodModeActionDTO struct {
	Type    GodModeActionTypes `json:"Type"`
	Payload []byte             `json:"Payload"`
}

func UnmarshalGodModeAction(raw []byte) (GodModeAction, error) {
	var rgma GodModeActionDTO
	if err := json.Unmarshal(raw, &rgma); err != nil {
		return nil, fmt.Errorf("failed to unmarshall god mode action: %v", err.Error())
	}

	var gma GodModeAction
	switch rgma.Type {
	case TypePlayPause:
		gma = &PlayPause{}
	case TypeInterfaceChange:
		gma = &InterfaceChange{}
	case TypeDispatch:
		gma = &Dispatch{}
	default:
		return nil, fmt.Errorf("failed to unmarshal god mode action. Unknown action type: %s", rgma.Type)
	}

	if err := json.Unmarshal(rgma.Payload, gma); err != nil {
		return nil, fmt.Errorf("failed to unmarshal god mode action of type %s: %v", rgma.Type, err)
	}

	return gma, nil

}
