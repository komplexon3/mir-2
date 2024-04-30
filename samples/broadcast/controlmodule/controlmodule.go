package controlmodule

import (
	"bufio"
	"context"
	"fmt"
	"os"

	es "github.com/go-errors/errors"
	"github.com/google/uuid"

	"github.com/filecoin-project/mir/pkg/modules"

	broadcastevents "github.com/filecoin-project/mir/samples/broadcast/events"

	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"
)

type controlModule struct {
	eventsOut    chan *stdtypes.EventList
	moduleConfig ControlModuleConfig
}

type ControlModuleConfig struct {
	Self      stdtypes.ModuleID // id of this module
	Broadcast stdtypes.ModuleID
}

func NewControlModule(mc ControlModuleConfig) modules.ActiveModule {
	return &controlModule{
		eventsOut:    make(chan *stdtypes.EventList),
		moduleConfig: mc,
	}
}

func (m *controlModule) ImplementsModule() {}

func (m *controlModule) ApplyEvents(_ context.Context, events *stdtypes.EventList) error {
	iter := events.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {

		switch evt := event.(type) {
		case *stdevents.Init:
			go func() {
				for {
					err := m.readMessageFromConsole()
					if err != nil {
						panic(err)
					}
				}
			}()
		case *broadcastevents.Deliver:
			fmt.Println("\nBroadcast says: ", string(evt.Data))
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
		broadcastevents.NewBroadcastRequest(m.moduleConfig.Broadcast, []byte(scanner.Text()), uuid.New()),
	)

	return nil
}
