package control

import (
	"bufio"
	"context"
	"fmt"
	"os"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/modules"
	broadcastevents "github.com/filecoin-project/mir/samples/authenticated-broadcast/events"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"
)

type controlModule struct {
	eventsOut chan *stdtypes.EventList
	isLeader  bool
}

func NewControlModule(isLeader bool) modules.ActiveModule {
	return &controlModule{
		eventsOut: make(chan *stdtypes.EventList),
		isLeader:  isLeader,
	}
}

func (m *controlModule) ImplementsModule() {}

func (m *controlModule) ApplyEvents(_ context.Context, events *stdtypes.EventList) error {
	iter := events.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {
		switch evt := event.(type) {
		case *stdevents.Init:
			if m.isLeader {
				go func() {
					err := m.readMessageFromConsole()
					if err != nil {
						panic(err)
					}
				}()
			} else {
				fmt.Println("Waiting for the message...")
			}
		case *broadcastevents.Deliver:
			fmt.Println("Leader says: ", string(evt.Data))
		default:
			return es.Errorf("unknown event type: %T", event)
		}
	}

	return nil
}

func (m *controlModule) EventsOut() <-chan *stdtypes.EventList {
	return m.eventsOut
}

func (m *controlModule) readMessageFromConsole() error {
	// Read the user input
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Type in a message and press Enter: ")
	scanner.Scan()
	if scanner.Err() != nil {
		return es.Errorf("error reading from console: %w", scanner.Err())
	}

	m.eventsOut <- stdtypes.ListOf(
		broadcastevents.NewBroadcastRequest("broadcast", scanner.Text()),
	)

	return nil
}
