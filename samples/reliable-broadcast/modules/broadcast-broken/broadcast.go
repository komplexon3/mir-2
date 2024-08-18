package broadcast

import (
	es "github.com/go-errors/errors"
	"github.com/pkg/errors"

	"github.com/filecoin-project/mir/samples/reliable-broadcast/events"
	"github.com/filecoin-project/mir/samples/reliable-broadcast/messages"
	"github.com/filecoin-project/mir/stdevents"
	eventsdsl "github.com/filecoin-project/mir/stdevents/dsl"
	"github.com/filecoin-project/mir/stdtypes"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
)

// ModuleConfig sets the module ids. All replicas are expected to use identical module configurations.
type ModuleConfig struct {
	Self     stdtypes.ModuleID // id of this module
	Consumer stdtypes.ModuleID // id of the module to send the "Deliver" event to
	Net      stdtypes.ModuleID
}

// ModuleParams sets the values for the parameters of an instance of the protocol.
// All replicas are expected to use identical module parameters.
type ModuleParams struct {
	InstanceUID []byte            // unique identifier for this instance of BCB, used to prevent cross-instance replay attacks
	AllNodes    []stdtypes.NodeID // the list of participating nodes
	Leader      stdtypes.NodeID   // the id of the leader of the instance
}

// GetN returns the total number of nodes.
func (params *ModuleParams) GetN() int {
	return len(params.AllNodes)
}

// GetF returns the maximum tolerated number of faulty nodes.
func (params *ModuleParams) GetF() int {
	return (params.GetN() - 1) / 3
}

// broadcastModuleState represents the state of the broadcast module.
type broadcastModuleState struct {
	// this variable is not part of the original protocol description, but it greatly simplifies the code
	request string

	sentEcho      bool
	sentReady     bool
	delivered     bool
	receivedEcho  map[stdtypes.NodeID]bool
	receivedReady map[stdtypes.NodeID]bool
}

// NewModule returns a passive module for the Signed Echo Broadcast from the textbook "Introduction to reliable and
// secure distributed programming". It serves as a motivating example for the DSL module interface.
// The pseudocode can also be found in https://dcl.epfl.ch/site/_media/education/sdc_byzconsensus.pdf (Algorithm 4
// (Echo broadcast [Rei94]))
func NewModule(mc ModuleConfig, params *ModuleParams, nodeID stdtypes.NodeID, logger logging.Logger) modules.PassiveModule {
	m := dsl.NewModule(mc.Self)

	state := broadcastModuleState{
		request: "",

		sentEcho:      false,
		sentReady:     false,
		delivered:     false,
		receivedEcho:  make(map[stdtypes.NodeID]bool),
		receivedReady: make(map[stdtypes.NodeID]bool),
	}

	dsl.UponEvent(m, func(br *events.BroadcastRequest) error {
		data := br.Data
		if nodeID != params.Leader {
			return es.Errorf("only the leader node can receive requests")
		}

		eventsdsl.SendMessage(m, mc.Net, mc.Self, messages.NewStartMessage(data), params.AllNodes...)
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
			if me.Sender == params.Leader && state.request == "" {
				state.request = msg.Data
				state.sentEcho = true
				eventsdsl.SendMessage(m, mc.Net, mc.Self, messages.NewEchoMessage(state.request), params.AllNodes...)
			}
			return nil
		case *messages.EchoMessage:
			if !state.receivedEcho[me.Sender] && msg.Data == state.request {
				state.receivedEcho[me.Sender] = true
			}
			return nil
		case *messages.ReadyMessage:
			if !state.receivedReady[me.Sender] && msg.Data == state.request {
				state.receivedReady[me.Sender] = true
			}
			return nil

		}
		logger.Log(logging.LevelWarn, "Reveived message with unknown payload type", "payload", me.Payload)
		return nil
	})

	// upon exists m != ⊥ such that #({p ∈ Π | echos[p] = m}) > (N+f)/2 and sentfinal = FALSE do
	dsl.UponStateUpdates(m, func() error {
		if !state.sentReady && len(state.receivedEcho) > (params.GetN()+params.GetF())/2 {
			state.sentReady = true
			eventsdsl.SendMessage(m, mc.Net, mc.Self, messages.NewReadyMessage(state.request), params.AllNodes...)
		}
		if !state.sentReady && len(state.receivedReady) > params.GetF() {
			state.sentReady = true
			eventsdsl.SendMessage(m, mc.Net, mc.Self, messages.NewReadyMessage(state.request), params.AllNodes...)
		}
		if !state.delivered && len(state.receivedReady) > 2*params.GetF() {
			state.delivered = true
			dsl.EmitEvent(m, events.NewDeliver(mc.Consumer, state.request))
		}
		return nil
	})

	return m
}
