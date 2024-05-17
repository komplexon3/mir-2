package debugger

import (
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/filecoin-project/mir/pkg/debugger/server"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/stdtypes"
)

type DebugWriter struct {
	server *server.Server
	logger logging.Logger
}

type DebugAction struct {
	Type  string `json:"Type"`
	Value string `json:"Value"`
}

// newDebugWriter creates a new WSWriter that establishes a websocket connection
func newDebugWriter(port string, logger logging.Logger) *DebugWriter {
	server := server.NewServer(port, logger)
	go server.StartServer()

	return &DebugWriter{
		server: server,
		logger: logger,
	}

}

// Write sends every event to the frontend and then waits for a message detailing how to proceed with that event
// The returned EventList contains the accepted events
func (wsw *DebugWriter) Write(list *stdtypes.EventList, timestamp int64) (*stdtypes.EventList, error) {
	wsw.logger.Log(logging.LevelInfo, "Writing events to interface")

	for !wsw.server.HasWSConnections() {
		wsw.logger.Log(logging.LevelInfo, "Waiting interface connection to proceed")
		time.Sleep(time.Millisecond * 1000)
	}

	if list.Len() == 0 {
		return list, nil
	}

	acceptedEvents := stdtypes.EmptyList()
	iter := list.Iterator()

	for event := iter.Next(); event != nil; event = iter.Next() {
		// skip non-protobuf events
		eventPb, ok := event.(*eventpb.Event)
		if !ok {
			acceptedEvents.PushBack(event)
			continue
		}

		eventJSON, err := protojson.Marshal(eventPb)
		if err != nil {
			return list, fmt.Errorf("error marshaling event to JSON: %w", err)
		}
		message, err := json.Marshal(map[string]interface{}{
			"event":     string(eventJSON),
			"timestamp": timestamp,
		})
		if err != nil {
			return list, fmt.Errorf("error marshaling eventJSON and timestamp to JSON: %w", err)
		}

		// TODO: should be method of Server
		wsw.server.SendEvent(message)

		actionMessage := wsw.server.ReceiveEventAction()
		// TODO: re-add the  logic from the old code regarding connection closing
		// try to unmarshal the action
		var action DebugAction
		err = json.Unmarshal(actionMessage.Payload, &action)
		if err != nil {
			return list, fmt.Errorf("error unmarshalling action to map: %w", err)
		}

		acceptedEvents, err = eventAction(action.Type, action.Value, acceptedEvents, eventPb)
		if err != nil {
			return list, err
		}
	}
	return acceptedEvents, nil
}

func (wsw *DebugWriter) Close() error {
	wsw.server.Close()
	return nil // TODO: should probably actually thing about errors
}

// Note: just to satisfy the interface, this method does nothing
func (wsw *DebugWriter) Flush() error {
	return nil
}

// EventAction decides, based on the input what exactly is done next with the current event
func eventAction(
	actionType string,
	value string,
	acceptedEvents *stdtypes.EventList,
	currentEvent *eventpb.Event,
) (*stdtypes.EventList, error) {
	if actionType == "accept" {
		acceptedEvents.PushBack(currentEvent)
	} else if actionType == "replace" {
		type ValueFormat struct {
			EventJSON string `json:"event"`
			Timestamp int64  `json:"timestamp"`
		}
		var input ValueFormat
		err := json.Unmarshal([]byte(value), &input)
		if err != nil {
			return acceptedEvents, fmt.Errorf("error unmarshalling value to ValueFormat: %w", err)
		}
		var modifiedEvent eventpb.Event
		err = protojson.Unmarshal([]byte(input.EventJSON), &modifiedEvent)
		if err != nil {
			return acceptedEvents, fmt.Errorf("error unmarshalling modified event using protojson: %w", err)
		}
		acceptedEvents.PushBack(&modifiedEvent)
	}
	return acceptedEvents, nil
}
