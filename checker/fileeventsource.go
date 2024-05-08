package checker

import (
	"os"
	"slices"

	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/recordingpb"
	"github.com/filecoin-project/mir/stdtypes"
)

// NOTE: if we deal with biiig logs, we should probably start streaming...
// TODO: this is not general purpose, need a better way to decode serialized events

func GetEventsFromFile(serializedDecoders []func(sEvt []byte) (stdtypes.Event, error), filenames ...string) []stdtypes.Event {
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
				for _, decoder := range serializedDecoders {
					if decoded, err := decoder(serializedEvent.Serialized.GetData()); err == nil {
						event = decoded
						break
					}
				}
			}
			eventLog = append(eventLog, event)
			// event.SetMetadata("node", stdtypes.NodeID(rle.NodeId))
		}
	}

	return eventLog
}
