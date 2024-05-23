package adversary

import (
	"runtime/debug"

	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

/** TODO: make interface which only needs to support "ProcessEvents"
 * Using a module would be one of many ways to implement it
 * Maybe add a wrapper where u just pass a module and it uses it as a plugin.
 */

type Plugin struct {
	name   string
	module modules.PassiveModule
}

func NewPlugin(name string, module modules.PassiveModule) *Plugin {
	return &Plugin{
		name:   name,
		module: module,
	}
}

func (p *Plugin) processEvents(evts *stdtypes.EventList) (outEvts *stdtypes.EventList, err error) {
	defer func() {
		if r := recover(); r != nil {
			if rErr, ok := r.(error); ok {
				err = es.Errorf("module panicked: %w\nStack trace:\n%s", rErr, string(debug.Stack()))
			} else {
				err = es.Errorf("module panicked: %v\nStack trace:\n%s", r, string(debug.Stack()))
			}
		}
	}()

	return p.module.ApplyEvents(evts)
}
