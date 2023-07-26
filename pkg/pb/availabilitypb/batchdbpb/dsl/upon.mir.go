// Code generated by Mir codegen. DO NOT EDIT.

package batchdbpbdsl

import (
	types2 "github.com/filecoin-project/mir/pkg/availability/multisigcollector/types"
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	types4 "github.com/filecoin-project/mir/pkg/trantor/types"
)

// Module-specific dsl functions for processing events.

func UponEvent[W types.Event_TypeWrapper[Ev], Ev any](m dsl.Module, handler func(ev *Ev) error) {
	dsl.UponMirEvent[*types1.Event_BatchDb](m, func(ev *types.Event) error {
		w, ok := ev.Type.(W)
		if !ok {
			return nil
		}

		return handler(w.Unwrap())
	})
}

func UponLookupBatch(m dsl.Module, handler func(batchId types2.BatchID, origin *types.LookupBatchOrigin) error) {
	UponEvent[*types.Event_Lookup](m, func(ev *types.LookupBatch) error {
		return handler(ev.BatchId, ev.Origin)
	})
}

func UponLookupBatchResponse[C any](m dsl.Module, handler func(found bool, txs []*types3.Transaction, context *C) error) {
	UponEvent[*types.Event_LookupResponse](m, func(ev *types.LookupBatchResponse) error {
		originWrapper, ok := ev.Origin.Type.(*types.LookupBatchOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.Found, ev.Txs, context)
	})
}

func UponStoreBatch(m dsl.Module, handler func(batchId types2.BatchID, txs []*types3.Transaction, retentionIndex types4.RetentionIndex, origin *types.StoreBatchOrigin) error) {
	UponEvent[*types.Event_Store](m, func(ev *types.StoreBatch) error {
		return handler(ev.BatchId, ev.Txs, ev.RetentionIndex, ev.Origin)
	})
}

func UponBatchStored[C any](m dsl.Module, handler func(context *C) error) {
	UponEvent[*types.Event_Stored](m, func(ev *types.BatchStored) error {
		originWrapper, ok := ev.Origin.Type.(*types.StoreBatchOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(context)
	})
}

func UponGarbageCollect(m dsl.Module, handler func(retentionIndex types4.RetentionIndex) error) {
	UponEvent[*types.Event_GarbageCollect](m, func(ev *types.GarbageCollect) error {
		return handler(ev.RetentionIndex)
	})
}
