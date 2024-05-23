package utilinterceptors

import (
	"fmt"

	"github.com/filecoin-project/mir/stdtypes"
)

type EventPrinter struct {
	NodeID stdtypes.NodeID
}

// TODO: extend to support non-broadcast events
func (ep *EventPrinter) Intercept(events *stdtypes.EventList) (*stdtypes.EventList, error) {
	iter := events.Iterator()
	for ev := iter.Next(); ev != nil; ev = iter.Next() {
		fmt.Printf("%s - %v\n\n", ep.NodeID, ev.ToString())
	}

	return events, nil
}
