package bcb

import (
	"fmt"

	es "github.com/go-errors/errors"
	"github.com/pkg/errors"

	"github.com/filecoin-project/mir/samples/bcb-native/events"
	"github.com/filecoin-project/mir/samples/bcb-native/messages"
	"github.com/filecoin-project/mir/stdevents"
	eventsdsl "github.com/filecoin-project/mir/stdevents/dsl"
	"github.com/filecoin-project/mir/stdtypes"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	cryptopbdsl "github.com/filecoin-project/mir/pkg/pb/cryptopb/dsl"
	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

//TODO Sanitize messages received by this module (e.g. check that the sender is the expected one, make sure no crashing if data=nil, etc.)

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

// bcbModuleState represents the state of the bcb module.
type bcbModuleState struct {
	// this variable is not part of the original protocol description, but it greatly simplifies the code
	request []byte

	sentEcho     bool
	sentFinal    bool
	delivered    bool
	receivedEcho map[stdtypes.NodeID]bool
	echoSigs     map[stdtypes.NodeID][]byte
}

// NewModule returns a passive module for the Signed Echo Broadcast from the textbook "Introduction to reliable and
// secure distributed programming". It serves as a motivating example for the DSL module interface.
// The pseudocode can also be found in https://dcl.epfl.ch/site/_media/education/sdc_byzconsensus.pdf (Algorithm 4
// (Echo broadcast [Rei94]))
func NewModule(mc ModuleConfig, params *ModuleParams, nodeID stdtypes.NodeID, logger logging.Logger) modules.PassiveModule {
	m := dsl.NewModule(mc.Self)

	state := bcbModuleState{
		request: nil,

		sentEcho:     false,
		sentFinal:    false,
		delivered:    false,
		receivedEcho: make(map[stdtypes.NodeID]bool),
		echoSigs:     make(map[stdtypes.NodeID][]byte),
	}

	dsl.UponEvent(m, func(br *events.BroadcastRequest) error {
		data := br.Data
		if nodeID != params.Leader {
			return es.Errorf("only the leader node can receive requests")
		}
		state.request = data

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
			if me.Sender == params.Leader && !state.sentEcho {
				sigMsg := &cryptopbtypes.SignedData{Data: [][]byte{params.InstanceUID, []byte("ECHO"), msg.Data}}
				cryptopbdsl.SignRequest(m, mc.Crypto, sigMsg, &signStartMessageContext{})
			}
			return nil
		case *messages.EchoMessage:
			if nodeID == params.Leader && !state.receivedEcho[me.Sender] && state.request != nil {
				state.receivedEcho[me.Sender] = true
				sigMsg := &cryptopbtypes.SignedData{Data: [][]byte{params.InstanceUID, []byte("ECHO"), state.request}}
				cryptopbdsl.VerifySig(m, mc.Crypto, sigMsg, msg.Signature, me.Sender, &verifyEchoContext{msg.Signature})
			}
			return nil
		case *messages.FinalMessage:
			if len(msg.Signers) == len(msg.Signatures) && len(msg.Signers) > (params.GetN()+params.GetF())/2 && !state.delivered {
				signedMessage := [][]byte{params.InstanceUID, []byte("ECHO"), msg.Data}
				sigMsgs := sliceutil.Repeat(&cryptopbtypes.SignedData{Data: signedMessage}, len(msg.Signers))
				cryptopbdsl.VerifySigs(m, mc.Crypto, sigMsgs, msg.Signatures, msg.Signers, &verifyFinalContext{msg.Data})
			}
			return nil

		}
		logger.Log(logging.LevelWarn, "Reveived message with unknown payload type", "payload", me.Payload)
		return nil
	})

	cryptopbdsl.UponSignResult(m, func(signature []byte, _ *signStartMessageContext) error {
		if !state.sentEcho {
			state.sentEcho = true
			echoMsg := messages.NewEchoMessage(signature)

			eventsdsl.SendMessage(m, mc.Net, mc.Self, echoMsg, params.Leader)
		}
		return nil
	})

	cryptopbdsl.UponSigVerified(m, func(nodeID stdtypes.NodeID, err error, context *verifyEchoContext) error {
		if err == nil {
			state.echoSigs[nodeID] = context.signature
		}
		return nil
	})

	// upon exists m != ⊥ such that #({p ∈ Π | echos[p] = m}) > (N+f)/2 and sentfinal = FALSE do
	dsl.UponStateUpdates(m, func() error {
		if len(state.echoSigs) > (params.GetN()+params.GetF())/2 && !state.sentFinal {
			state.sentFinal = true
			certSigners, certSignatures := maputil.GetKeysAndValues(state.echoSigs)
			eventsdsl.SendMessage(m, mc.Net, mc.Self, messages.NewFinalMessage(state.request, certSigners, certSignatures), params.AllNodes...)
		}
		return nil
	})

	cryptopbdsl.UponSigsVerified(m, func(_ []stdtypes.NodeID, _ []error, allOK bool, context *verifyFinalContext) error {
		if allOK && !state.delivered {
			state.delivered = true
			fmt.Println("Deliver")
			dsl.EmitEvent(m, events.NewDeliver(mc.Consumer, context.data))
		}
		return nil
	})

	return m
}

// Context data structures

type signStartMessageContext struct{}

type verifyEchoContext struct {
	signature []byte
}

type verifyFinalContext struct {
	data []byte
}
