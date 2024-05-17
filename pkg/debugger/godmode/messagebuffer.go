package godmode

import (
	"time"

	"github.com/filecoin-project/mir/stdtypes"
)

// Note: for now, we will simply make it a slice based queue. Later, doubly linked might make sense
// Also, without the possiblility of injecting events, we will only buffer one element anyways

type MessageBufferElement struct {
	Message   stdtypes.Message
	Timestamp time.Time
}

type MessageBuffer struct {
	queue []*MessageBufferElement
}

func (mb *MessageBuffer) PushMessage(m stdtypes.Message) {
	mbe := MessageBufferElement{
		Message:   m,
		Timestamp: time.Now(),
	}
	mb.queue = append(mb.queue, &mbe)
}

func (mb *MessageBuffer) PopMessage() *MessageBufferElement {
	if mb.Size() == 0 {
		return nil
	}

	var mbe *MessageBufferElement
	mbe, mb.queue = mb.queue[0], mb.queue[1:]
	return mbe
}

func (mb *MessageBuffer) Size() int {
	return len(mb.queue)
}
