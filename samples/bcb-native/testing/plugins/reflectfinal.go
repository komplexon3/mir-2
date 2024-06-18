package plugins

import (
	"github.com/filecoin-project/mir/adversary"
	"github.com/filecoin-project/mir/pkg/dsl"
	broadcastmessages "github.com/filecoin-project/mir/samples/broadcast/messages"
	"github.com/filecoin-project/mir/stdevents"
	eventsdsl "github.com/filecoin-project/mir/stdevents/dsl"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

func NewReflectFinalPlugin(transportModule stdtypes.ModuleID, allNonByzantineNodes []stdtypes.NodeID) *adversary.Plugin {
	m := dsl.NewModule("reflectFinal")

	dsl.UponEvent(m, func(me *stdevents.MessageReceived) error {
		// pass through original event)
		dsl.EmitEvent(m, me)

		data, err := me.Payload.ToBytes()
		if err != nil {
			return es.Errorf("could not retrieve data from received message: %v", err)
		}

		msgRaw, err := broadcastmessages.Deserialize(data)
		if err != nil {
			return es.Errorf("could not deserialize message from received message data: %v", err)
		}

		switch msg := msgRaw.(type) {
		case *broadcastmessages.FinalMessage:
			// reflect to all nodes but itself
			eventsdsl.SendMessage(m, transportModule, me.DestModule, msg, allNonByzantineNodes...)
			// dsl.EmitEvent(m, stdevents.NewSendMessage(transportModule, me.DestModule, msg, allNonByzantineNodes...))
		}

		// _ = broadcastmessages.EchoMessageMsg
		// _ = es.Errorf("")
		return nil
	})

	dsl.UponOtherEvent(m, func(e stdtypes.Event) error {
		dsl.EmitEvent(m, e)
		return nil
	})

	return adversary.NewPlugin("reflectFinal", m)
}
