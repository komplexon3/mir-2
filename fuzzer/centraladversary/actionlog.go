package centraladversay

import (
	"fmt"

	"github.com/filecoin-project/mir/stdtypes"
)

type actionTraceEntry struct {
	node      stdtypes.NodeID
	actionLog string
}

type actionTrace struct {
	history []actionTraceEntry
}

func newActionTrace() *actionTrace {
	return &actionTrace{
		history: make([]actionTraceEntry, 0),
	}
}

func (a *actionTrace) String() string {
	str := ""
	for _, al := range a.history {
		str += fmt.Sprintf("Node %s took action: %v\n\n", al.node, al.actionLog)
	}
	return str
}

func (a *actionTrace) Push(node stdtypes.NodeID, actionLog string) {
	a.history = append(a.history, actionTraceEntry{node, actionLog})
}
