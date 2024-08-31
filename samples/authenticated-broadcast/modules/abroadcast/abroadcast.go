package abroadcast

import (
	es "github.com/go-errors/errors"
	"github.com/pkg/errors"

	"github.com/filecoin-project/mir/samples/authenticated-broadcast/events"
	"github.com/filecoin-project/mir/samples/authenticated-broadcast/messages"
	"github.com/filecoin-project/mir/stdevents"
	eventsdsl "github.com/filecoin-project/mir/stdevents/dsl"
	"github.com/filecoin-project/mir/stdtypes"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
)

// TODO Sanitize messages received by this module (e.g. check that the sender is the expected one, make sure no crashing if data=nil, etc.)

// ModuleConfig sets the module ids. All replicas are expected to use identical module configurations.
type ModuleConfig struct {
	Self     stdtypes.ModuleID // id of this module
	Consumer stdtypes.ModuleID // id of the module to send the "Deliver" event to
	Net      stdtypes.ModuleID
	Crypto   stdtypes.ModuleID
}

// ModuleParams sets the values for the parameters of an instance of the protocol.
// All replicas are expected to use identical module parameters.
type ModuleParams struct {
	AllNodes []stdtypes.NodeID // the list of participating nodes
	Leader   stdtypes.NodeID   // the id of the leader of the instance
}

// GetN returns the total number of nodes.
func (params *ModuleParams) GetN() int {
	return len(params.AllNodes)
}

// GetF returns the maximum tolerated number of faulty nodes.
func (params *ModuleParams) GetF() int {
	return (params.GetN() - 1) / 3
}

// bcbModuleState represents the state of the bcb module.
type bcbModuleState struct {
	delivered    bool
	sentEcho     bool
	receivedEcho map[string]map[stdtypes.NodeID]struct{} // used as set
}

// NewModule returns a passive module for the Signed Echo Broadcast from the textbook "Introduction to reliable and
// secure distributed programming". It serves as a motivating example for the DSL module interface.
// The pseudocode can also be found in https://dcl.epfl.ch/site/_media/education/sdc_byzconsensus.pdf (Algorithm 4
// (Echo broadcast [Rei94]))
func NewModule(mc ModuleConfig, params *ModuleParams, nodeID stdtypes.NodeID, logger logging.Logger) modules.PassiveModule {
	m := dsl.NewModule(mc.Self)

	state := bcbModuleState{
		delivered:    false,
		sentEcho:     false,
		receivedEcho: make(map[string]map[stdtypes.NodeID]struct{}),
	}

	dsl.UponEvent(m, func(br *events.BroadcastRequest) error {
		if nodeID != params.Leader {
			return es.Errorf("only the leader node can receive requests")
		}

		eventsdsl.SendMessage(m, mc.Net, mc.Self, messages.NewStartMessage(br.Data), params.AllNodes...)
		return nil
	})

	dsl.UponEvent(m, func(me *stdevents.MessageReceived) error {
		data, err := me.Payload.ToBytes()
		if err != nil {
			return errors.Errorf("could not retrieve data from received message: %v", err)
		}

		msgRaw, err := messages.Deserialize(data)
		if err != nil {
			return errors.Errorf("could not deserialize message from received message data: %v", err)
		}

		switch msg := msgRaw.(type) {
		case *messages.StartMessage:
			if me.Sender == params.Leader {
				state.sentEcho = true
				echoMsg := messages.NewEchoMessage(msg.Data)
				eventsdsl.SendMessage(m, mc.Net, mc.Self, echoMsg, params.AllNodes...)
			}
			return nil
		case *messages.EchoMessage:
			if state.receivedEcho[msg.Data] == nil {
				state.receivedEcho[msg.Data] = make(map[stdtypes.NodeID]struct{})
			}
			state.receivedEcho[msg.Data][me.Sender] = struct{}{}
			return nil
		}
		logger.Log(logging.LevelWarn, "Reveived message with unknown payload type", "payload", me.Payload)
		return nil
	})

	dsl.UponStateUpdates(m, func() error {
		if state.delivered {
			return nil
		}
		for req, echos := range state.receivedEcho {
			if len(echos) > (params.GetN()+params.GetF())/2 {
				state.delivered = true
				dsl.EmitEvent(m, events.NewDeliver(mc.Consumer, req))
				return nil
			}
		}
		return nil
	})

	return m
}