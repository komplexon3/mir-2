package sharedkeysmodule

import (
	"bytes"

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

// Idea: all byz nodes share their private keys with each other
// byz sender uses all keys two sign two different versions of the messages
// good nodes are split in two, one gets a start message for the first message and the other one for the second one
// -> send final messages for different messages to the good nodes (byz nodes other than the sender don't do nothing and will not deliver anything)

// ModuleConfig sets the module ids. All replicas are expected to use identical module configurations.
type ModuleConfig struct {
	Self     stdtypes.ModuleID // id of this module
	Consumer stdtypes.ModuleID // id of the module to send the "Deliver" event to
	Net      stdtypes.ModuleID
	Crypto   map[stdtypes.NodeID]stdtypes.ModuleID
}

// ModuleParams sets the values for the parameters of an instance of the protocol.
// All replicas are expected to use identical module parameters.
type ModuleParams struct {
	NodeID         stdtypes.NodeID
	AllNodes       []stdtypes.NodeID // the list of participating nodes
	ByzantineNodes []stdtypes.NodeID
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

	instances       map[uuid.UUID]*bcbInstanceState
	attackInstances map[uuid.UUID]*bcbAttackInstance

	moduleConfig ModuleConfig
	params       ModuleParams
	logger       logging.Logger

	primaryGoodNodes   []stdtypes.NodeID
	secondaryGoodNodes []stdtypes.NodeID
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
	broadcastID       uuid.UUID
	primaryData       []byte
	primaryEchoSigs   map[stdtypes.NodeID][]byte
	secondaryData     []byte
	secondaryEchoSigs map[stdtypes.NodeID][]byte
	sentFinal         bool
}

func newBcbAttackInstance(primaryBroadcastID uuid.UUID, primaryData []byte) *bcbAttackInstance {
	secondaryData := append(primaryData, primaryData...)

	return &bcbAttackInstance{
		broadcastID:       primaryBroadcastID,
		primaryData:       primaryData,
		primaryEchoSigs:   make(map[stdtypes.NodeID][]byte),
		secondaryData:     secondaryData,
		secondaryEchoSigs: make(map[stdtypes.NodeID][]byte),
		sentFinal:         false,
	}
}

func NewBroadcast(mp ModuleParams, mc ModuleConfig, logger logging.Logger) dsl.Module {
	m := dsl.NewModule(mc.Self)

	goodNodes := sliceutil.Filter(mp.AllNodes, func(_ int, t stdtypes.NodeID) bool {
		return !sliceutil.Contains(mp.ByzantineNodes, t)
	})

	// TODO: setup broadcast
	b := Broadcast{
		m:               m,
		instances:       make(map[uuid.UUID]*bcbInstanceState),
		attackInstances: make(map[uuid.UUID]*bcbAttackInstance),
		moduleConfig:    mc,
		params:          mp,
		logger:          logger,

		primaryGoodNodes:   goodNodes[:len(goodNodes)/2],
		secondaryGoodNodes: goodNodes[len(goodNodes)/2:],
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
  b.logger.Log(logging.LevelWarn, "BYZANTINE")
	return nil
}

func (b *Broadcast) handleBroadcastRequest(e *events.BroadcastRequest) error {
	// TODO: add new entry, emil start message
	instance := newBcbAttackInstance(e.BroadcastID, e.Data) // NOTE: byzantine node could set arbitrary sender
	b.attackInstances[e.BroadcastID] = instance

	primaryStartMessage := messages.NewStartMessage(instance.primaryData, b.params.NodeID, instance.broadcastID)
	secondaryStartMessage := messages.NewStartMessage(instance.secondaryData, b.params.NodeID, instance.broadcastID)

	b.SendMessage(primaryStartMessage, b.primaryGoodNodes...)
	b.SendMessage(secondaryStartMessage, b.secondaryGoodNodes...)

	// create double signatures for all byzantine nodes

	for nodeID, cryptoID := range b.moduleConfig.Crypto {
		primarySigMsg := &cryptopbtypes.SignedData{Data: [][]byte{instance.broadcastID[:], []byte("ECHO"), instance.primaryData}}
		cryptopbdsl.SignRequest(b.m, cryptoID, primarySigMsg, &signStartMessageContext{instance.broadcastID, instance.primaryData, nodeID})
		secondarySigMsg := &cryptopbtypes.SignedData{Data: [][]byte{instance.broadcastID[:], []byte("ECHO"), instance.secondaryData}}
		cryptopbdsl.SignRequest(b.m, cryptoID, secondarySigMsg, &signStartMessageContext{instance.broadcastID, instance.secondaryData, nodeID})
	}

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
		sigMsg = &cryptopbtypes.SignedData{Data: [][]byte{instance.broadcastID[:], []byte("ECHO"), instance.data}}
		cryptopbdsl.SignRequest(b.m, b.moduleConfig.Crypto[b.params.NodeID], sigMsg, &signStartMessageContext{instance.broadcastID, instance.data, b.params.NodeID})
	}

	return nil
}

func (b *Broadcast) handleSignResponse(signature []byte, c *signStartMessageContext) error {
	if attack, ok := b.attackInstances[c.broadcastID]; ok {
		// it is one of our attacks
		if bytes.Equal(attack.primaryData, c.data) {
			attack.primaryEchoSigs[c.signerID] = signature
		} else {
			attack.secondaryEchoSigs[c.signerID] = signature
		}

		return nil
	}

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
		b.logger.Log(logging.LevelWarn, "Received echo for unknown broadcast or one that this node did not initiate.")
		return nil
	}

  if sliceutil.Contains(b.primaryGoodNodes, from) {
		sigMsg := &cryptopbtypes.SignedData{Data: [][]byte{instance.broadcastID[:], []byte("ECHO"), instance.primaryData}}
		cryptopbdsl.VerifySig(b.m, b.moduleConfig.Crypto[b.params.NodeID], sigMsg, e.Signature, from, &verifyEchoContext{instance.broadcastID, from, e.Signature})
  } else {
		sigMsg := &cryptopbtypes.SignedData{Data: [][]byte{instance.broadcastID[:], []byte("ECHO"), instance.secondaryData}}
		cryptopbdsl.VerifySig(b.m, b.moduleConfig.Crypto[b.params.NodeID], sigMsg, e.Signature, from, &verifyEchoContext{instance.broadcastID, from, e.Signature})
  }

	return nil
}

func (b *Broadcast) handleSigVerified(nodeID stdtypes.NodeID, err error, c *verifyEchoContext) error {
	instance, ok := b.attackInstances[c.broadcastID]
	if !ok {
		b.logger.Log(logging.LevelError, "SigVerified for unknown broadcast id", "broadcastID", c.broadcastID)
	}

	if err != nil {
    return nil
  }

  if sliceutil.Contains(b.primaryGoodNodes, c.signerID) {
		instance.primaryEchoSigs[nodeID] = c.signature
	} else {
		instance.secondaryEchoSigs[nodeID] = c.signature
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
		cryptopbdsl.VerifySigs(b.m, b.moduleConfig.Crypto[b.params.NodeID], sigMsgs, e.Signatures, e.Signers, &verifyFinalContext{instance.broadcastID})
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

func (b *Broadcast) checkEmitFinal() error {
	for _, instance := range b.attackInstances {
		if len(instance.primaryEchoSigs) > (b.params.GetN()+b.params.GetF())/2&&len(instance.secondaryEchoSigs) > (b.params.GetN()+b.params.GetF())/2  && !instance.sentFinal {
			instance.sentFinal = true
		pCertSigners, pCertSignatures := maputil.GetKeysAndValues(instance.primaryEchoSigs)
			b.SendMessage(messages.NewFinalMessage(instance.primaryData, b.params.NodeID, instance.broadcastID, pCertSigners, pCertSignatures), b.primaryGoodNodes...)
		sCertSigners, sCertSignatures := maputil.GetKeysAndValues(instance.secondaryEchoSigs)
			b.SendMessage(messages.NewFinalMessage(instance.secondaryData, b.params.NodeID, instance.broadcastID, sCertSigners, sCertSignatures), b.secondaryGoodNodes...)
		}
	}

	return nil
}

// Context data structures
type signStartMessageContext struct {
	broadcastID uuid.UUID
	data        []byte
	signerID    stdtypes.NodeID
}

type verifyEchoContext struct {
	broadcastID uuid.UUID
  signerID stdtypes.NodeID
	signature   []byte
}

type verifyFinalContext struct {
	broadcastID uuid.UUID
}
