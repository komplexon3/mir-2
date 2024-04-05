package godmode

import (
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/filecoin-project/mir/pkg/debugger/godmode/types"
	"github.com/filecoin-project/mir/pkg/debugger/server"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"
)

type GodModeWriter struct {
	server *server.Server
	logger logging.Logger
	state  types.NodeInterfaceState
}

// newWSWriter creates a new WSWriter that establishes a websocket connection
func newWSWriter(port string, logger logging.Logger) *GodModeWriter {
	server := server.NewServer(port, logger)
	go server.StartServer()

	return &GodModeWriter{
		server: server,
		logger: logger,
		state:  types.Blocking,
	}

}

// Write sends every event to the frontend and then waits for a message detailing how to proceed with that event
// The returned EventList contains the accepted events
func (wsw *GodModeWriter) Write(list *stdtypes.EventList, timestamp int64) (*stdtypes.EventList, error) {
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
		var message stdtypes.Message
		switch evt := event.(type) {
		case *stdevents.SendMessage:
			message = evt.Payload
		case *stdevents.MessageReceived:
			message = evt.Payload
		default:
			wsw.logger.Log(logging.LevelError, "Network God Mode Interceptor received non-message event. This should not happen. Event: %s", evt.ToString())

		}

		switch wsw.state {
		case types.Open:
			acceptedEvents.PushBack(event)
			msgEvent := types.NewMessageEvent(message, types.TypeMessageDispatched, time.Now())
			payload, err := msgEvent.Marshal()
			if err != nil {
				wsw.logger.Log(logging.LevelError, "Failed to marshal message. Just letting it pass anyways. Error: %v", err)
			}
			wsw.server.SendGodModeMessage(payload)
		case types.Dropping:
			// not adding it to the accepted events
			msgEvent := types.NewMessageEvent(message, types.TypeMessageDropped, time.Now())
			payload, err := msgEvent.Marshal()
			if err != nil {
				wsw.logger.Log(logging.LevelError, "Failed to marshal message. Just letting it pass anyways. Error: %v", err)
			}
			wsw.server.SendGodModeMessage(payload)
		case types.Blocking:
			// not adding it to the accepted events
			msgEvent := types.NewMessageEvent(message, types.TypeMesssageBuffered, time.Now())
			payload, err := msgEvent.Marshal()
			if err != nil {
				wsw.logger.Log(logging.LevelError, "Failed to marshal message. Just letting it pass anyways. Error: %v", err)
			}
			wsw.server.SendGodModeMessage(payload)

			// Blocking until action taken
			actionRaw := wsw.server.ReceiveGodModeMessage()
			action := types.UnmarshalGodModeAction()
		}

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

		// actionMessage := wsw.server.ReceiveAction()
		// TODO: re-add the  logic from the old code regarding connection closing
		// try to unmarshal the action
		// var action DebugAction
		// err = json.Unmarshal(actionMessage.Payload, &action)
		// if err != nil {
		// 	return list, fmt.Errorf("error unmarshalling action to map: %w", err)
		// }

		// acceptedEvents, err = eventAction(action.Type, action.Value, acceptedEvents, eventPb)
		// if err != nil {
		// 	return list, err
		// }
	}
	return acceptedEvents, nil
}

func (wsw *GodModeWriter) Close() error {
	wsw.server.Close()
	return nil // TODO: should probably actually thing about errors
}

// Note: just to satisfy the interface, this method does nothing
func (wsw *GodModeWriter) Flush() error {
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
