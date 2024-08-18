package deepcopymsgs

import (
	"github.com/brunoga/deep"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

type DeepCopyMsgs struct{}

func (i *DeepCopyMsgs) Intercept(events *stdtypes.EventList) (*stdtypes.EventList, error) {
	newEvts := stdtypes.EmptyList()
	evtIterator := events.Iterator()
	for e := evtIterator.Next(); e != nil; e = evtIterator.Next() {
		if sendMsg, ok := e.(*stdevents.SendMessage); ok {
			sendMsgCopy, err := deep.Copy(sendMsg)
			if err != nil {
				return nil, es.Errorf("failed to create deep copy of send message: %v", err)
			}
			newEvts.PushBack(sendMsgCopy)
		} else {
			newEvts.PushBack(e)
		}
	}
	return newEvts, nil
}
