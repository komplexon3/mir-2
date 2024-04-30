package utilinterceptors

import (
	"github.com/filecoin-project/mir/stdtypes"
)

type NodeIdMetadataInterceptor struct {
	NodeID stdtypes.NodeID
}

// TODO: extend to support non-broadcast events
func (i *NodeIdMetadataInterceptor) Intercept(events *stdtypes.EventList) (*stdtypes.EventList, error) {
  newEvents := stdtypes.EmptyList()
	iter := events.Iterator()
	for ev := iter.Next(); ev != nil; ev = iter.Next() {
      nEv, err := ev.SetMetadata("node", i.NodeID)
      if err != nil {
        return nil, err
      }
      newEvents.PushBack(nEv)
		}

	return events, nil
}
