package byzmodule

import (
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
	NodeID         stdtypes.NodeID
	AllNodes       []stdtypes.NodeID // the list of participating nodes
}

// GetN returns the total number of nodes.
func (params *ModuleParams) GetN() int {
	return len(params.AllNodes)
}

// GetF returns the maximum tolerated number of faulty nodes.
func (params *ModuleParams) GetF() int {
	return (params.GetN() - 1) / 3
}

type Broadcast struct {
	m dsl.Module

	instances map[uuid.UUID]*bcbInstanceState
	// mapped from primary and secondary broadcastId
	attackInstances map[uuid.UUID]*bcbAttackInstance

	moduleConfig ModuleConfig
	params       ModuleParams
	logger       logging.Logger

	primaryNodes   []stdtypes.NodeID
	secondaryNodes []stdtypes.NodeID
	mixedNodes     []stdtypes.NodeID
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

type bcbAttackInstance struct {
	primaryBroadcastId   uuid.UUID
	primaryData          []byte
	primaryEchoSigs      map[stdtypes.NodeID][]byte
	secondaryBroadcastId uuid.UUID
	secondaryData        []byte
	secondaryEchoSigs    map[stdtypes.NodeID][]byte
	sentFinal            bool
}

func newBcbAttackInstance(primaryBroadcastId uuid.UUID, primaryData []byte) *bcbAttackInstance {
	secondaryBroadcastId := uuid.New()
	secondaryData := append(primaryData, primaryData...)

	return &bcbAttackInstance{
		primaryBroadcastId:   primaryBroadcastId,
		primaryData:          primaryData,
		primaryEchoSigs:      make(map[stdtypes.NodeID][]byte),
		secondaryBroadcastId: secondaryBroadcastId,
		secondaryData:        secondaryData,
		secondaryEchoSigs:    make(map[stdtypes.NodeID][]byte),
		sentFinal:            false,
	}
}

func NewBroadcast(mp ModuleParams, mc ModuleConfig, logger logging.Logger) dsl.Module {
	m := dsl.NewModule(mc.Self)
  
	// TODO: setup broadcast
	b := Broadcast{
		m:               m,
		instances:       make(map[uuid.UUID]*bcbInstanceState),
		attackInstances: make(map[uuid.UUID]*bcbAttackInstance),
		moduleConfig:    mc,
		params:          mp,
		logger:          logger,

		mixedNodes:   mp.AllNodes[:(len(mp.AllNodes) / 3)],
		primaryNodes: mp.AllNodes[(len(mp.AllNodes) / 3):(2*len(mp.AllNodes)/3)],
    secondaryNodes: mp.AllNodes[(2*len(mp.AllNodes)/3):],

	}

	dsl.UponEvent(m, func(_ *stdevents.Init) error {
		return b.handleInit()
	})

	dsl.UponEvent(m, b.handleBroadcastRequest)

	cryptopbdsl.UponSignResult(m, b.handleSignResponse)
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

	return m
}

func (b *Broadcast) SendMessage(msg stdtypes.Message, destNodes ...stdtypes.NodeID) {
	eventsdsl.SendMessage(b.m, b.moduleConfig.Net, b.moduleConfig.Self, msg, destNodes...)
}

func (b *Broadcast) handleInit() error {
  b.logger.Log(logging.LevelInfo,"starting up BYZANTINE")
	return nil
}

func (b *Broadcast) handleBroadcastRequest(e *events.BroadcastRequest) error {
	// TODO: add new entry, emil start message
	instance := newBcbAttackInstance(e.BroadcastID, e.Data)
	b.attackInstances[instance.primaryBroadcastId] = instance
	b.attackInstances[instance.secondaryBroadcastId] = instance

	primaryStartMessage := messages.NewStartMessage(instance.primaryData, b.params.NodeID, instance.primaryBroadcastId)
	mixedStartMessage := messages.NewStartMessage(instance.secondaryData, b.params.NodeID, instance.primaryBroadcastId)
	secondaryStartMessage := messages.NewStartMessage(instance.secondaryData, b.params.NodeID, instance.secondaryBroadcastId)

	b.SendMessage(primaryStartMessage, b.primaryNodes...)
	b.SendMessage(mixedStartMessage, b.secondaryNodes...)
	b.SendMessage(primaryStartMessage, b.mixedNodes...)
	b.SendMessage(secondaryStartMessage, b.mixedNodes...)

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
		var sigMsg *cryptopbtypes.SignedData
		sigMsg = &cryptopbtypes.SignedData{Data: [][]byte{[]byte("ECHO"), instance.data}}
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
	instance, ok := b.attackInstances[e.BroadcastID]

	if !ok {
		b.logger.Log(logging.LevelWarn, "Received echo for unknown broadcast (attack).")
		return nil
	}

	if sliceutil.Contains(b.primaryNodes, from) {
		instance.primaryEchoSigs[from] = e.Signature
	} else if sliceutil.Contains(b.secondaryNodes, from) {
		instance.secondaryEchoSigs[from] = e.Signature
	} else {
		// byz node
		if e.BroadcastID == instance.primaryBroadcastId {
			instance.primaryEchoSigs[from] = e.Signature
		} else {
			instance.secondaryEchoSigs[from] = e.Signature
		}
	}

	if len(instance.primaryEchoSigs) > (b.params.GetN()+b.params.GetF())/2 && len(instance.secondaryEchoSigs) > (b.params.GetN()+b.params.GetF())/2 && !instance.sentFinal {
		instance.sentFinal = true
		pCertSigners, pCertSignatures := maputil.GetKeysAndValues(instance.primaryEchoSigs)
		sCertSigners, sCertSignatures := maputil.GetKeysAndValues(instance.secondaryEchoSigs)

		b.SendMessage(messages.NewFinalMessage(instance.primaryData, b.params.NodeID, instance.primaryBroadcastId, pCertSigners, pCertSignatures), b.primaryNodes...)
		b.SendMessage(messages.NewFinalMessage(instance.secondaryData, b.params.NodeID, instance.primaryBroadcastId, sCertSigners, sCertSignatures), b.secondaryNodes...)

		b.SendMessage(messages.NewFinalMessage(instance.primaryData, b.params.NodeID, instance.primaryBroadcastId, pCertSigners, pCertSignatures), b.mixedNodes...)
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
		signedMessage := [][]byte{[]byte("ECHO"), instance.data}
		sigMsgs := sliceutil.Repeat(&cryptopbtypes.SignedData{Data: signedMessage}, len(e.Signers))
		cryptopbdsl.VerifySigs(b.m, b.moduleConfig.Crypto, sigMsgs, e.Signatures, e.Signers, &verifyFinalContext{instance.broadcastID})
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
		instance.delivered = true
		dsl.EmitEvent(b.m, events.NewDeliver(b.moduleConfig.Consumer, instance.data, instance.broadcastID, instance.senderID))
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

type verifyFinalContext struct {
	broadcastID uuid.UUID
}
