package plugins

import (
	"slices"

	"github.com/filecoin-project/mir/adversary"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/samples/broadcast/events"
	"github.com/filecoin-project/mir/samples/broadcast/messages"
	broadcastmessages "github.com/filecoin-project/mir/samples/broadcast/messages"
	"github.com/filecoin-project/mir/stdevents"
	eventsdsl "github.com/filecoin-project/mir/stdevents/dsl"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
	"github.com/google/uuid"
)

type req struct {
	id       uuid.UUID
	data     []byte
	echoSigs map[stdtypes.NodeID][]byte
}

type reqCombo struct {
	aReq req
	bReq req
}

type byzBrain struct {
	maliciousReqs map[uuid.UUID]*reqCombo
}

func NewDoubleBroadcast(goodNodes, byzantineNodes []stdtypes.NodeID, transportModule, broadcastModule stdtypes.ModuleID) *adversary.Plugin {
	m := dsl.NewModule("doubleBroadcast")

	n := len(goodNodes) + len(byzantineNodes)
	f := (n - 1) / 3

	aNodes := goodNodes[:len(goodNodes)/2-1]
	bNodes := goodNodes[len(goodNodes)/2:]

	brain := byzBrain{
		maliciousReqs: make(map[uuid.UUID]*reqCombo),
	}

	dsl.UponEvent(m, func(e *events.BroadcastRequest) error {
		nodeId, err := e.GetMetadata("node")
		if err != nil {
			return es.Errorf("failed to get node id from metadata: %v", err)
		}

		aId := e.BroadcastID
		aData := e.Data
		bId := uuid.New()
		bData := append(aData, aData...)

		combo := reqCombo{
			aReq: req{
				id:       aId,
				data:     bData,
				echoSigs: make(map[stdtypes.NodeID][]byte),
			},
			bReq: req{
				id:       bId,
				data:     bData,
				echoSigs: make(map[stdtypes.NodeID][]byte),
			},
		}

		brain.maliciousReqs[aId] = &combo
		brain.maliciousReqs[bId] = &combo

		msgA := messages.NewStartMessage(aData, nodeId.(stdtypes.NodeID), aId)
		msgBA := messages.NewStartMessage(bData, nodeId.(stdtypes.NodeID), aId)
		msgB := messages.NewStartMessage(bData, nodeId.(stdtypes.NodeID), bId)

		eventsdsl.SendMessage(m, transportModule, broadcastModule, msgA, byzantineNodes...)
		eventsdsl.SendMessage(m, transportModule, broadcastModule, msgB, byzantineNodes...)

		eventsdsl.SendMessage(m, transportModule, broadcastModule, msgA, aNodes...)
		eventsdsl.SendMessage(m, transportModule, broadcastModule, msgBA, bNodes...)

		return nil
	})

	dsl.UponEvent(m, func(me *stdevents.MessageReceived) error {

		data, err := me.Payload.ToBytes()
		if err != nil {
			return es.Errorf("could not retrieve data from received message: %v", err)
		}

		msgRaw, err := broadcastmessages.Deserialize(data)
		if err != nil {
			return es.Errorf("could not deserialize message from received message data: %v", err)
		}

		switch msg := msgRaw.(type) {
		case *broadcastmessages.EchoMessage:
			doubleReq, ok := brain.maliciousReqs[msg.BroadcastID]
			if !ok {
				dsl.EmitEvent(m, me)
				return nil
			}
			if slices.Contains(aNodes, me.Sender) {
				doubleReq.aReq.echoSigs[me.Sender] = msg.Signature
				return nil
			} else if slices.Contains(bNodes, me.Sender) {
				doubleReq.bReq.echoSigs[me.Sender] = msg.Signature
				return nil
			} else {
				if doubleReq.aReq.id == msg.BroadcastID {
					doubleReq.aReq.echoSigs[me.Sender] = msg.Signature
				} else {
					doubleReq.bReq.echoSigs[me.Sender] = msg.Signature
				}

			}

			_nodeId, err := me.GetMetadata("node")
			if err != nil {
				return es.Errorf("failed to get node id from metadata: %v", err)
			}
			nodeId := stdtypes.NodeID(_nodeId.(string))
			if len(doubleReq.aReq.echoSigs) > (n+f)/2 {
				certSigners, certSignatures := maputil.GetKeysAndValues(doubleReq.aReq.echoSigs)
				eventsdsl.SendMessage(m, transportModule, broadcastModule, messages.NewFinalMessage(doubleReq.aReq.data, nodeId, doubleReq.aReq.id, certSigners, certSignatures), aNodes...)
			}

			if len(doubleReq.bReq.echoSigs) > (n+f)/2 {
				certSigners, certSignatures := maputil.GetKeysAndValues(doubleReq.bReq.echoSigs)
				eventsdsl.SendMessage(m, transportModule, broadcastModule, messages.NewFinalMessage(doubleReq.bReq.data, nodeId, doubleReq.bReq.id, certSigners, certSignatures), bNodes...)
			}
		default:
			dsl.EmitEvent(m, me)

		}

		return nil
	})

	dsl.UponOtherEvent(m, func(e stdtypes.Event) error {
		dsl.EmitEvent(m, e)
		return nil
	})

	return adversary.NewPlugin("doubleBroadcast", m)
}
