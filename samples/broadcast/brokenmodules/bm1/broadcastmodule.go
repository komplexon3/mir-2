package bm1

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/filecoin-project/mir/samples/broadcast/events"
	"github.com/filecoin-project/mir/samples/broadcast/messages"
	"github.com/filecoin-project/mir/stdevents"
	eventsdsl "github.com/filecoin-project/mir/stdevents/dsl"
	"github.com/filecoin-project/mir/stdtypes"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	cryptopbdsl "github.com/filecoin-project/mir/pkg/pb/cryptopb/dsl"
	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

/**
 * IDEA: Provoke double deliver by having one node (say 1) re-transmit final message.
 * All nodes need to fail to set "delivered".
 * => Failes a possible interpretation of integrity.
 *
 * another way to break it (for bm4):
* the handle final doesn't verify that the signatures are from unqiue nodes... but broadcast id is checked :/
 *
 * All changes are marked with // CHANGE
*/

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
	NodeID   stdtypes.NodeID
	AllNodes []stdtypes.NodeID // the list of participating nodes
}

// GetN returns the total number of nodes.
func (params *ModuleParams) GetN() int {
	return len(params.AllNodes)
}

// GetF returns the maximum tolerated number of faulty nodes.
func (params *ModuleParams) GetF() int {
	return (params.GetN() - 1) / 2
}

type Broadcast struct {
	m dsl.Module

	instances    map[uuid.UUID]*bcbInstanceState
	ourInstances []*bcbInstanceState

	moduleConfig ModuleConfig
	params       ModuleParams
	logger       logging.Logger
}

// bcbInstanceState represents the state of one instance of bcb.
// NOTE: not coordinating access to these - risk of race conditions?
type bcbInstanceState struct {
	broadcastID uuid.UUID
	senderID    stdtypes.NodeID
	data        []byte

	sentEcho     bool
	sentFinal    bool
	delivered    bool
	receivedEcho map[stdtypes.NodeID]bool
	echoSigs     map[stdtypes.NodeID][]byte
}

func newBcbInstance(senderID stdtypes.NodeID, broadcastID uuid.UUID, data []byte) *bcbInstanceState {
	return &bcbInstanceState{
		broadcastID: broadcastID,
		senderID:    senderID,
		data:        data,

		sentEcho:     false,
		sentFinal:    false,
		delivered:    false,
		receivedEcho: make(map[stdtypes.NodeID]bool),
		echoSigs:     make(map[stdtypes.NodeID][]byte),
	}
}

func NewBroadcast(mp ModuleParams, mc ModuleConfig, logger logging.Logger) dsl.Module {
	m := dsl.NewModule(mc.Self)

	// TODO: setup broadcast
	b := Broadcast{
		m:            m,
		instances:    make(map[uuid.UUID]*bcbInstanceState),
		ourInstances: make([]*bcbInstanceState, 0),
		moduleConfig: mc,
		params:       mp,
		logger:       logger,
	}

	dsl.UponEvent(m, func(_ *stdevents.Init) error {
		return b.handleInit()
	})

	dsl.UponEvent(m, b.handleBroadcastRequest)

	cryptopbdsl.UponSignResult(m, b.handleSignResponse)
	cryptopbdsl.UponSigVerified(m, b.handleSigVerified)
	cryptopbdsl.UponSigsVerified(m, b.handleSigsVerified)

	// TODO: can this be handled in a 'dsl-way'?
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
			return b.handleStartMessage(me.Sender, msg)
		case *messages.EchoMessage:
			return b.handleEchoMessage(me.Sender, msg)
		case *messages.FinalMessage:
			return b.handleFinalMessage(me.Sender, msg)
		}

		b.logger.Log(logging.LevelWarn, "Reveived message with unknown payload type", "payload", me.Payload)

		return nil
	})

	dsl.UponStateUpdate(b.m, b.checkEmitFinal)

	return m
}

func (b *Broadcast) SendMessage(msg stdtypes.Message, destNodes ...stdtypes.NodeID) {
	eventsdsl.SendMessage(b.m, b.moduleConfig.Net, b.moduleConfig.Self, msg, destNodes...)
}

func (b *Broadcast) handleInit() error {
	return nil
}

func (b *Broadcast) handleBroadcastRequest(e *events.BroadcastRequest) error {
	instance := newBcbInstance(b.params.NodeID, e.BroadcastID, e.Data) // NOTE: byzantine node could set arbitrary sender
	b.instances[e.BroadcastID] = instance
	b.ourInstances = append(b.ourInstances, instance)

	startMessage := messages.NewStartMessage(instance.data, instance.senderID, instance.broadcastID)

	b.SendMessage(startMessage, b.params.AllNodes...)

	return nil
}

func (b *Broadcast) handleStartMessage(_ stdtypes.NodeID, e *messages.StartMessage) error {
	// should be unkown to us unless we created the request
	instance, ok := b.instances[e.BroadcastID]

	if !ok {
		// NOTE: not checking that the message is actually comming from the specified broadcast sender
		instance = newBcbInstance(e.BroadcastSender, e.BroadcastID, e.Data)
		b.instances[e.BroadcastID] = instance
	}

	if !instance.sentEcho {
		sigMsg := &cryptopbtypes.SignedData{Data: [][]byte{instance.broadcastID[:], []byte("ECHO"), instance.data}}
		cryptopbdsl.SignRequest(b.m, b.moduleConfig.Crypto, sigMsg, &signStartMessageContext{instance.broadcastID})
	}

	return nil
}

func (b *Broadcast) handleSignResponse(signature []byte, c *signStartMessageContext) error {
	instance, ok := b.instances[c.broadcastID]
	if !ok {
		b.logger.Log(logging.LevelError, "SignResult for unknown broadcast id", "broadcastID", c.broadcastID)
	}

	if !instance.sentEcho {
		instance.sentEcho = true
		echoMsg := messages.NewEchoMessage(signature, c.broadcastID)
		b.SendMessage(echoMsg, instance.senderID)
	}

	return nil
}

func (b *Broadcast) handleEchoMessage(from stdtypes.NodeID, e *messages.EchoMessage) error {
	instance, ok := b.instances[e.BroadcastID]

	if !ok || instance.senderID != b.params.NodeID {
		b.logger.Log(logging.LevelWarn, "Received echo for unknown broadcast or one that this node did not initiate.")
		return nil
	}

	// other possibility for bug (allow for one node to submit multiple sigs)
	if !instance.receivedEcho[from] && instance.data != nil {
		instance.receivedEcho[from] = true
		sigMsg := &cryptopbtypes.SignedData{Data: [][]byte{instance.broadcastID[:], []byte("ECHO"), instance.data}}
		cryptopbdsl.VerifySig(b.m, b.moduleConfig.Crypto, sigMsg, e.Signature, from, &verifyEchoContext{instance.broadcastID, e.Signature})
	}

	return nil
}

func (b *Broadcast) handleSigVerified(nodeID stdtypes.NodeID, err error, c *verifyEchoContext) error {
	instance, ok := b.instances[c.broadcastID]
	if !ok {
		b.logger.Log(logging.LevelError, "SigVerified for unknown broadcast id", "broadcastID", c.broadcastID)
	}

	if err == nil {
		instance.echoSigs[nodeID] = c.signature
	}
	return nil
}

func (b *Broadcast) handleFinalMessage(from stdtypes.NodeID, e *messages.FinalMessage) error {
	instance, ok := b.instances[e.BroadcastID]
	if !ok {
		b.logger.Log(logging.LevelWarn, "Received final for unknown broadcast")
		return nil
	}

	// here we can provoke 'double deliver'
	if len(e.Signers) == len(e.Signatures) && len(e.Signers) > (b.params.GetN()+b.params.GetF())/2 && !instance.delivered {
		signedMessage := [][]byte{instance.broadcastID[:], []byte("ECHO"), instance.data}
		sigMsgs := sliceutil.Repeat(&cryptopbtypes.SignedData{Data: signedMessage}, len(e.Signers))
		// CHANGE
		// add signatures, signers to context
		cryptopbdsl.VerifySigs(b.m, b.moduleConfig.Crypto, sigMsgs, e.Signatures, e.Signers, &verifyFinalContext{instance.broadcastID, e.Signatures, e.Signers})
	}

	return nil
}

func (b *Broadcast) handleSigsVerified(_ []stdtypes.NodeID, _ []error, allOK bool, c *verifyFinalContext) error {
	instance, ok := b.instances[c.broadcastID]
	if !ok {
		b.logger.Log(logging.LevelError, "SigsVerified for unknown broadcast id", "broadcastID", c.broadcastID)
	}

	// here we can provoke 'double deliver'
	if allOK && !instance.delivered {
		// CHANGE
		// disable setting delivered to true
		// instance.delivered = true
		dsl.EmitEvent(b.m, events.NewDeliver(b.moduleConfig.Consumer, instance.data, instance.broadcastID, instance.senderID))
	}

	// CHANGE
	// node 1 will retransmit final message once
	// after one retransmit, setting delivered to prevent infinite retransmit
	if b.params.NodeID == stdtypes.NodeID("1") && !instance.delivered {
		instance.delivered = true
		fmt.Println("\nByzantine retransmit final message")
		b.SendMessage(messages.NewFinalMessage(instance.data, instance.senderID, instance.broadcastID, c.signers, c.signatures), b.params.AllNodes...)
	}

	return nil

}

func (b *Broadcast) checkEmitFinal() error {
	for _, instance := range b.ourInstances {

		if len(instance.echoSigs) > (b.params.GetN()+b.params.GetF())/2 && !instance.sentFinal {
			instance.sentFinal = true
			certSigners, certSignatures := maputil.GetKeysAndValues(instance.echoSigs)
			b.SendMessage(messages.NewFinalMessage(instance.data, instance.senderID, instance.broadcastID, certSigners, certSignatures), b.params.AllNodes...)
		}
	}

	return nil
}

// Context data structures
type signStartMessageContext struct {
	broadcastID uuid.UUID
}

type verifyEchoContext struct {
	broadcastID uuid.UUID
	signature   []byte
}

// CHANGE
// add signatures, signers to context
type verifyFinalContext struct {
	broadcastID uuid.UUID
	signatures  [][]byte
	signers     []stdtypes.NodeID
}
