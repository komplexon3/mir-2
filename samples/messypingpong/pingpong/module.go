package pingpong

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/filecoin-project/mir/samples/messypingpong/events"
	"github.com/filecoin-project/mir/stdevents"
	stddsl "github.com/filecoin-project/mir/stdevents/dsl"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
)

type ModuleConfig struct {
	Self      stdtypes.ModuleID // id of this module
	Other     stdtypes.ModuleID
	Transport stdtypes.ModuleID
	SelfNode  stdtypes.NodeID
	OtherNode stdtypes.NodeID
}

type ppm struct {
	m      dsl.Module
	logger logging.Logger
	mc     ModuleConfig
}

func NewModule(mc ModuleConfig, logger logging.Logger) modules.PassiveModule {
	m := dsl.NewModule(mc.Self)

	p := ppm{
		m,
		logger,
		mc,
	}

	dsl.UponEvent(m, func(_ *stdevents.Init) error {
		if mc.Self == stdtypes.ModuleID("A") {
			pp := events.NewPingPong(mc.SelfNode, mc.Self, 0)
			p.sendPingPong(pp)
		}
		logger.Log(logging.LevelTrace, "Initiating ping pong - sent ping pong", "module", mc.Self)
		return nil
	})

	dsl.UponEvent(m, func(pp *events.PingPong) error {
		return p.handlePingPong(pp)
	})

	dsl.UponEvent(m, func(msg *stdevents.MessageReceived) error {
		logger.Log(logging.LevelTrace, "Received message", "sender", msg.Sender, "msg", msg.Payload)
		data, err := msg.Payload.ToBytes()
		if err != nil {
			return es.Errorf("could not retrieve data from received message: %v", err)
		}

		msgRaw, err := events.Deserialize(data)
		if err != nil {
			return es.Errorf("could not deserialize message from received message data: %v", err)
		}
		switch pp := msgRaw.(type) {
		case *events.PingPong:
			return p.handlePingPong(pp)
		default:
			return es.Errorf("unknown msg/event: %T", pp)
		}
	})

	return m
}

func (p *ppm) handlePingPong(pp *events.PingPong) error {

	const reps uint64 = 1
	if pp.SeqNr >= reps {
		p.logger.Log(logging.LevelTrace, fmt.Sprintf("%d rounds over, lets chill", reps), "module", p.mc.Self)
		return nil
	}

	p.logger.Log(logging.LevelTrace, "Received ping pong", "module", p.mc.Self)
	time.Sleep(time.Second)
	p.sendPingPong(pp.Next())
	return nil
}

func (p *ppm) sendPingPong(pp *events.PingPong) {
	// send to this node or the node, if going to the same node, 50/50
	stddsl.SendMessage(p.m, p.mc.Transport, p.mc.Self, pp, p.mc.OtherNode)
	p.logger.Log(logging.LevelTrace, "Sent ping pong", "srcModule", p.mc.Self, "destNode", p.mc.OtherNode, "destModule", pp.Dest())

	if rand.Float32() > 0.5 {
		// or "send" to this node
		p.logger.Log(logging.LevelTrace, "Sent ping pong to internally", "srcModule", p.mc.Self, "destModule", pp.Dest())
		stddsl.SendMessage(p.m, p.mc.Transport, p.mc.Self, pp, p.mc.SelfNode)
	} else {
		// or deliver directly to this node
		p.logger.Log(logging.LevelTrace, "Sent ping pong as event", "srcModule", p.mc.Self, "destModule", pp.Dest())
		dsl.EmitEvent(p.m, pp)
	}

	// if rand.Float32() > 0.5 {
	// 	// send to another node
	// 	stddsl.SendMessage(p.m, p.mc.Transport, p.mc.Self, pp, p.mc.OtherNode)
	// 	p.logger.Log(logging.LevelTrace, "Sent ping pong", "srcModule", p.mc.Self, "destNode", p.mc.OtherNode, "destModule", pp.Dest())
	// } else if rand.Float32() > 0.5 {
	// 	// or "send" to this node
	// 	p.logger.Log(logging.LevelTrace, "Sent ping pong to internally", "srcModule", p.mc.Self, "destModule", pp.Dest())
	// 	stddsl.SendMessage(p.m, p.mc.Transport, p.mc.Self, pp, p.mc.SelfNode)
	// } else {
	// 	// or deliver directly to this node
	// 	p.logger.Log(logging.LevelTrace, "Sent ping pong as event", "srcModule", p.mc.Self, "destModule", pp.Dest())
	// 	dsl.EmitEvent(p.m, pp)
	// }
}
