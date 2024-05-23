package plugins

import (
	"github.com/filecoin-project/mir/adversary"
	"github.com/filecoin-project/mir/pkg/dsl"
	broadcastmessages "github.com/filecoin-project/mir/samples/broadcast/messages"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

func NewNoEcho() *adversary.Plugin {
	m := dsl.NewModule("noEcho")

	dsl.UponEvent(m, func(me *stdevents.MessageReceived) error {

		data, err := me.Payload.ToBytes()
		if err != nil {
			return es.Errorf("could not retrieve data from received message: %v", err)
		}

		msgRaw, err := broadcastmessages.Deserialize(data)
		if err != nil {
			return es.Errorf("could not deserialize message from received message data: %v", err)
		}

		switch msgRaw.(type) {
		case *broadcastmessages.StartMessage:
			// ignore them (i.e., don't echo)
			return nil
		}

		// just pass through all other
		dsl.EmitEvent(m, me)
		return nil
	})

	dsl.UponOtherEvent(m, func(e stdtypes.Event) error {
		dsl.EmitEvent(m, e)
		return nil
	})

	return adversary.NewPlugin("noEcho", m)
}
