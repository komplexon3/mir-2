package checker

import (
	"os"
	"slices"

	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/recordingpb"
	broadcastevents "github.com/filecoin-project/mir/samples/broadcast/events"
	"github.com/filecoin-project/mir/stdtypes"
)

// NOTE: if we deal with biiig logs, we should probably start streaming...

func GetEventsFromFile(filenames ...string) []stdtypes.Event {
	recordingLog := make([]*recordingpb.Entry, 0)

	for _, fn := range filenames {
		file, err := os.Open(fn)
		if err != nil {
			panic(err)
		}
		defer file.Close() // ignore close errors

		reader, err := eventlog.NewReader(file)
		if err != nil {
			panic(err)
		}

		entries, err := reader.ReadAllEntries()
		if err != nil {
			panic(err)
		}

		recordingLog = append(recordingLog, entries...)
	}

	// actually, we want to interleave the histories
	// for simplicity we join them and then sort them

	// keeping original order of items
	slices.SortStableFunc(recordingLog, func(a, b *recordingpb.Entry) int {
		return int(a.GetTime() - b.GetTime())
	})

	eventLog := make([]stdtypes.Event, 0, len(recordingLog))
	for _, rle := range recordingLog {
		for _, e := range rle.Events {
			var event stdtypes.Event = e
			if serializedEvent, ok := event.(*eventpb.Event).Type.(*eventpb.Event_Serialized); ok {
				decodedEvent, err := broadcastevents.Deserialize(serializedEvent.Serialized.GetData())
				if err != nil {
					// TODO: why is (almost?) every event a serialized event?
					// fmt.Println("failed to deserialize event, just passing it on...")
				} else {
					event = decodedEvent
				}
			}
			eventLog = append(eventLog, event)
			// event.SetMetadata("node", stdtypes.NodeID(rle.NodeId))
		}
	}

	return eventLog
}
