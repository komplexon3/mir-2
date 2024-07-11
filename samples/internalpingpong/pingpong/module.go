package pingpong

import (
	"time"

	"github.com/filecoin-project/mir/samples/internalpingpong/events"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
)

type ModuleConfig struct {
	Self  stdtypes.ModuleID // id of this module
	Other stdtypes.ModuleID
}

func NewModule(mc ModuleConfig, logger logging.Logger) modules.PassiveModule {
	m := dsl.NewModule(mc.Self)

	dsl.UponEvent(m, func(_ *stdevents.Init) error {
		if m.ModuleID() == "A" {
			dsl.EmitEvent(m, events.NewPingPong(mc.Other, 0))
			logger.Log(logging.LevelTrace, "Initiating ping pong - sent ping pong", "module", mc.Self)
		} else {
			logger.Log(logging.LevelTrace, "Initiating ping pong - not A module, waiting on ping pong", "module", mc.Self)
		}
		return nil
	})

	dsl.UponEvent(m, func(pp *events.PingPong) error {
		// sleep 1 sec
		if pp.SeqNr >= 10 {
			logger.Log(logging.LevelTrace, "10 rounds over, lets chill", "module", mc.Self)
			return nil
		}

		logger.Log(logging.LevelTrace, "Received ping pong", "module", mc.Self)
		time.Sleep(time.Second)
		dsl.EmitEvent(m, events.NewPingPong(mc.Other, pp.SeqNr+1))
		logger.Log(logging.LevelTrace, "Sent ping pong", "module", mc.Self)
		return nil
	})
	return m
}
