// Code generated by Mir codegen. DO NOT EDIT.

package bcmpb

import (
	reflect "reflect"
)

func (*Event) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*Event_NewBlock)(nil)),
		reflect.TypeOf((*Event_NewChain)(nil)),
		reflect.TypeOf((*Event_GetChainRequest)(nil)),
		reflect.TypeOf((*Event_GetChainResponse)(nil)),
		reflect.TypeOf((*Event_GetHeadToCheckpointChainRequest)(nil)),
		reflect.TypeOf((*Event_GetHeadToCheckpointChainResponse)(nil)),
		reflect.TypeOf((*Event_RegisterCheckpoint)(nil)),
		reflect.TypeOf((*Event_InitBlockchain)(nil)),
	}
}
