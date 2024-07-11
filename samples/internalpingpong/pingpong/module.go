package pingpong

import (
	"math/rand/v2"
	"time"

	"github.com/filecoin-project/mir/samples/internalpingpong/events"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
)

type ModuleConfig struct {
	Self   stdtypes.ModuleID // id of this module
	Others []stdtypes.ModuleID
}

func NewModule(mc ModuleConfig, logger logging.Logger) modules.PassiveModule {
	m := dsl.NewModule(mc.Self)

	dsl.UponEvent(m, func(_ *stdevents.Init) error {
		destModule := mc.Others[rand.IntN(len(mc.Others))]
		dsl.EmitEvent(m, events.NewPingPong(destModule, 0))
		logger.Log(logging.LevelTrace, "Initiating ping pong - sent ping pong", "module", mc.Self)
		return nil
	})

	dsl.UponEvent(m, func(pp *events.PingPong) error {
		// sleep 1 sec
		if pp.SeqNr >= 5 {
			logger.Log(logging.LevelTrace, "5 rounds over, lets chill", "module", mc.Self)
			return nil
		}
		logger.Log(logging.LevelTrace, "Received ping pong", "module", mc.Self)
		time.Sleep(time.Second / 5)
		destModule := mc.Others[rand.IntN(len(mc.Others))]
		dsl.EmitEvent(m, events.NewPingPong(destModule, pp.SeqNr+1))
		logger.Log(logging.LevelTrace, "Sent ping pong", "module", mc.Self)
		return nil
	})
	return m
}
