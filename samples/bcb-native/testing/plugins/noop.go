package plugins

import (
	"github.com/filecoin-project/mir/adversary"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/stdtypes"
)

func NewNoOp() *adversary.Plugin {
	m := dsl.NewModule("noOp")

	dsl.UponOtherEvent(m, func(e stdtypes.Event) error {
		dsl.EmitEvent(m, e)
		return nil
	})

	return adversary.NewPlugin("noOp", m)
}
