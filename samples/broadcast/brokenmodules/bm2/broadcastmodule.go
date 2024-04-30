package bm2

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

/**
 * IDEA: Malicious broadcaster - break consistency by sending two different start messages to two messages.
 * After receiving all signatures, send the final message for one value to all but one node.
 * Send the final message for the other value to the remaining node.
 * Remaining node -> 3, Malicious broadcaster -> 0
 *
 * For this to work, the protocol is broken by not properly keeping track of the state of the broadcast instances.
 * The non-broadcasting nodes only take note of which broadcast instances have been delivered and who broadcasted it but not the value broadcasted
 * and whether or not the echo was sent.
 * Additionally, the broadcasting party doesn't use the map but only the array to keep track the bcbInstances
 * (NOTE: Why did I create this array? It is pretty useless... well... except for now...)
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
	return (params.GetN() - 1) / 3
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
	// CHANGE
	// create two instances, add both to array, send start message for both
	instance1 := newBcbInstance(b.params.NodeID, e.BroadcastID, []byte(string(e.Data)+"-1"))
	instance2 := newBcbInstance(b.params.NodeID, e.BroadcastID, []byte(string(e.Data)+"-2"))
	// CHANGE
	// b.instances[e.BroadcastID] = instance
	b.ourInstances = append(b.ourInstances, instance1, instance2)

	startMessage1 := messages.NewStartMessage(instance1.data, instance1.senderID, instance1.broadcastID)
	startMessage2 := messages.NewStartMessage(instance2.data, instance2.senderID, instance2.broadcastID)

	b.SendMessage(startMessage1, b.params.AllNodes...)
	b.SendMessage(startMessage2, b.params.AllNodes...)

	return nil
}

func (b *Broadcast) handleStartMessage(_ stdtypes.NodeID, e *messages.StartMessage) error {
	// should be unkown to us unless we created the request
	// instance, ok := b.instances[e.BroadcastID]
	//
	// if !ok {
	// 	// NOTE: not checking that the message is actually comming from the specified broadcast sender
	// 	instance = newBcbInstance(e.BroadcastSender, e.BroadcastID, e.Data)
	// 	b.instances[e.BroadcastID] = instance
	// }

	// CHANGE - don't keep track of actual state - just keep relevant info in context

	// NOTE: new instance registered - will forget about previous one
	// if !instance.sentEcho {
	sigMsg := &cryptopbtypes.SignedData{Data: [][]byte{e.BroadcastID[:], []byte("ECHO"), e.Data}}
	//
	cryptopbdsl.SignRequest(b.m, b.moduleConfig.Crypto, sigMsg, &signStartMessageContext{e.BroadcastID, e.BroadcastSender})
	// }

	return nil
}

func (b *Broadcast) handleSignResponse(signature []byte, c *signStartMessageContext) error {
	// instance, ok := b.instances[c.broadcastID]
	// if !ok {
	// 	b.logger.Log(logging.LevelError, "SignResult for unknown broadcast id", "broadcastID", c.broadcastID)
	// }
	//
	// if !instance.sentEcho {
	// instance.sentEcho = true
	echoMsg := messages.NewEchoMessage(signature, c.broadcastID)
	b.SendMessage(echoMsg, c.broadcastSender)
	// }

	return nil
}

func (b *Broadcast) handleEchoMessage(from stdtypes.NodeID, e *messages.EchoMessage) error {
	// CHANGE - just ignoring this, we will simply go though all the instances that we have and match in broadcast
	// instance, ok := b.instances[e.BroadcastID]
	//
	// if !ok || instance.senderID != b.params.NodeID {
	// 	b.logger.Log(logging.LevelWarn, "Received echo for unknown broadcast or one that this node did not initiate.")
	// 	return nil
	// }

	// if !instance.receivedEcho[from] && instance.data != nil {
	// 	instance.receivedEcho[from] = true
	// 	sigMsg := &cryptopbtypes.SignedData{Data: [][]byte{instance.broadcastID[:], []byte("ECHO"), instance.data}}
	// 	cryptopbdsl.VerifySig(b.m, b.moduleConfig.Crypto, sigMsg, e.Signature, from, &verifyEchoContext{instance.broadcastID, e.Signature})
	// }

	// CHANGE - simply sending for all matching broadcast id instances, only one will be successful - but that is okay
	for _, instance := range b.ourInstances {
		if instance.broadcastID != e.BroadcastID || instance.receivedEcho[from] {
			continue
		}
		sigMsg := &cryptopbtypes.SignedData{Data: [][]byte{instance.broadcastID[:], []byte("ECHO"), instance.data}}
		cryptopbdsl.VerifySig(b.m, b.moduleConfig.Crypto, sigMsg, e.Signature, from, &verifyEchoContext{instance.data, instance.broadcastID, e.Signature})
	}

	// if !instance.receivedEcho[from] && instance.data != nil {
	// 	instance.receivedEcho[from] = true
	// 	sigMsg := &cryptopbtypes.SignedData{Data: [][]byte{instance.broadcastID[:], []byte("ECHO"), instance.data}}
	// 	cryptopbdsl.VerifySig(b.m, b.moduleConfig.Crypto, sigMsg, e.Signature, from, &verifyEchoContext{instance.broadcastID, e.Signature})
	// }

	return nil
}

func (b *Broadcast) handleSigVerified(nodeID stdtypes.NodeID, err error, c *verifyEchoContext) error {
	// instance, ok := b.instances[c.broadcastID]
	// if !ok {
	// 	b.logger.Log(logging.LevelError, "SigVerified for unknown broadcast id", "broadcastID", c.broadcastID)
	// }
	//
	// if err == nil {
	// 	instance.echoSigs[nodeID] = c.signature
	// }

	// CHANGE - iterating through all instances to find mathing one
	// add signature and set "echo received from" (previouly ser in handleEcho
	for _, instance := range b.ourInstances {
		if instance.broadcastID != c.broadcastID || bytes.Equal(instance.data, c.data) || instance.receivedEcho[nodeID] {
			continue
		}
		instance.receivedEcho[nodeID] = true
		instance.echoSigs[nodeID] = c.signature
	}

	return nil
}

func (b *Broadcast) handleFinalMessage(from stdtypes.NodeID, e *messages.FinalMessage) error {
	// CHANGE - ignore instance data
	// instance, ok := b.instances[e.BroadcastID]
	// if !ok {
	// 	b.logger.Log(logging.LevelWarn, "Received final for unknown broadcast")
	// 	return nil
	// }

	// here we can provoke 'double deliver'

	data := e.Data
	broadcastID := e.BroadcastID

	if len(e.Signers) == len(e.Signatures) && len(e.Signers) > (b.params.GetN()+b.params.GetF())/2 {
		signedMessage := [][]byte{broadcastID[:], []byte("ECHO"), data}
		sigMsgs := sliceutil.Repeat(&cryptopbtypes.SignedData{Data: signedMessage}, len(e.Signers))
		cryptopbdsl.VerifySigs(b.m, b.moduleConfig.Crypto, sigMsgs, e.Signatures, e.Signers, &verifyFinalContext{data, broadcastID, e.BroadcastSender})
	}

	return nil
}

func (b *Broadcast) handleSigsVerified(_ []stdtypes.NodeID, _ []error, allOK bool, c *verifyFinalContext) error {
	// instance, ok := b.instances[c.broadcastID]
	// if !ok {
	// 	b.logger.Log(logging.LevelError, "SigsVerified for unknown broadcast id", "broadcastID", c.broadcastID)
	// }
	//

	// here we can provoke 'double deliver'
	// if allOK && !instance.delivered {
	if allOK {
		// instance.delivered = true
		dsl.EmitEvent(b.m, events.NewDeliver(b.moduleConfig.Consumer, c.data, c.broadcastID, c.broadcastSender))
	}
	return nil

}

func (b *Broadcast) checkEmitFinal() error {
	for _, instance := range b.ourInstances {
		if len(instance.echoSigs) > (b.params.GetN()+b.params.GetF())/2 && !instance.sentFinal {
			instance.sentFinal = true
			certSigners, certSignatures := maputil.GetKeysAndValues(instance.echoSigs)
			// data should never be empty with the -1 -2 postfixes
			// CHANGE - only send to some nodes. If ends with 1 send to all but node 3, node 3 gets the other value
			if string(instance.data[len(instance.data)-1:]) == "2" {
				b.SendMessage(messages.NewFinalMessage(instance.data, instance.senderID, instance.broadcastID, certSigners, certSignatures), stdtypes.NodeID("3"))
			} else {
				allNodesButNode3 := sliceutil.Filter(b.params.AllNodes, func(i int, t stdtypes.NodeID) bool { return t != stdtypes.NodeID("3") })
				b.SendMessage(messages.NewFinalMessage(instance.data, instance.senderID, instance.broadcastID, certSigners, certSignatures), allNodesButNode3...)
			}
		}
	}

	return nil
}

// Context data structures
// CHANGE - add sender
type signStartMessageContext struct {
	broadcastID     uuid.UUID
	broadcastSender stdtypes.NodeID
}

// CHANGE - add data
type verifyEchoContext struct {
	data        []byte // to differentiate between the multiple instances with the same broadcastID
	broadcastID uuid.UUID
	signature   []byte
}

// CHANGE - add data
type verifyFinalContext struct {
	data            []byte // to differentiate between the multiple instances with the same broadcastID
	broadcastID     uuid.UUID
	broadcastSender stdtypes.NodeID
}
