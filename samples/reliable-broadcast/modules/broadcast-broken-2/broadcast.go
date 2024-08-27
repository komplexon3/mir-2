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
	sentEcho      bool
	sentReady     bool
	delivered     bool
	receivedEcho  map[string]map[stdtypes.NodeID]struct{}
	receivedReady map[string]map[stdtypes.NodeID]struct{}
}

func NewModule(mc ModuleConfig, params *ModuleParams, nodeID stdtypes.NodeID, logger logging.Logger) modules.PassiveModule {
	m := dsl.NewModule(mc.Self)

	state := broadcastModuleState{
		sentEcho:      false,
		sentReady:     false,
		delivered:     false,
		receivedEcho:  make(map[string]map[stdtypes.NodeID]struct{}), // used as set
		receivedReady: make(map[string]map[stdtypes.NodeID]struct{}), // used as set
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
			if me.Sender == params.Leader {
				state.sentEcho = true
				eventsdsl.SendMessage(m, mc.Net, mc.Self, messages.NewEchoMessage(msg.Data), params.AllNodes...)
			}
			return nil
		case *messages.EchoMessage:
			if state.receivedEcho[msg.Data] == nil {
				state.receivedEcho[msg.Data] = make(map[stdtypes.NodeID]struct{})
			}
			state.receivedEcho[msg.Data][me.Sender] = struct{}{}
			return nil
		case *messages.ReadyMessage:
			if state.receivedReady[msg.Data] == nil {
				state.receivedReady[msg.Data] = make(map[stdtypes.NodeID]struct{})
			}
			state.receivedReady[msg.Data][me.Sender] = struct{}{}
			return nil

		}
		logger.Log(logging.LevelWarn, "Reveived message with unknown payload type", "payload", me.Payload)
		return nil
	})

	dsl.UponStateUpdates(m, func() error {
		for req, echos := range state.receivedEcho {
			if state.delivered {
				return nil
			}
			if !state.sentReady && len(echos) > (params.GetN()+params.GetF())/2 {
				state.sentReady = true
				eventsdsl.SendMessage(m, mc.Net, mc.Self, messages.NewReadyMessage(req), params.AllNodes...)
			}
		}

		for req, readies := range state.receivedReady {
			if !state.sentReady && len(readies) > params.GetF() {
				state.sentReady = true
				eventsdsl.SendMessage(m, mc.Net, mc.Self, messages.NewReadyMessage(req), params.AllNodes...)
			}

			if len(readies) > 2*params.GetF() {
				state.delivered = true
				dsl.EmitEvent(m, events.NewDeliver(mc.Consumer, req))
			}
		}
		return nil
	})

	return m
}
