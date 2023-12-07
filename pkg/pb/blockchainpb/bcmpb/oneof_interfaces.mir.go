// Code generated by Mir codegen. DO NOT EDIT.

package bcmpb

type Event_Type = isEvent_Type

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func (w *Event_NewBlock) Unwrap() *NewBlock {
	return w.NewBlock
}

func (w *Event_NewChain) Unwrap() *NewChain {
	return w.NewChain
}

func (w *Event_GetBlockRequest) Unwrap() *GetBlockRequest {
	return w.GetBlockRequest
}

func (w *Event_GetBlockResponse) Unwrap() *GetBlockResponse {
	return w.GetBlockResponse
}

func (w *Event_GetChainRequest) Unwrap() *GetChainRequest {
	return w.GetChainRequest
}

func (w *Event_GetChainResponse) Unwrap() *GetChainResponse {
	return w.GetChainResponse
}

func (w *Event_GetHeadToCheckpointChainRequest) Unwrap() *GetHeadToCheckpointChainRequest {
	return w.GetHeadToCheckpointChainRequest
}

func (w *Event_GetHeadToCheckpointChainResponse) Unwrap() *GetHeadToCheckpointChainResponse {
	return w.GetHeadToCheckpointChainResponse
}

func (w *Event_RegisterCheckpoint) Unwrap() *RegisterCheckpoint {
	return w.RegisterCheckpoint
}
