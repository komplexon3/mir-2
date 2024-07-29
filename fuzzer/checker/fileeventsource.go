package checker

import (
	"fmt"
	"os"
	"slices"

	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/recordingpb"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
	"github.com/filecoin-project/mir/pkg/vcinterceptor/vectorclock"
	"github.com/filecoin-project/mir/stdtypes"
)

// NOTE: if we deal with biiig logs, we should probably start streaming...
// TODO: this is not general purpose, need a better way to decode serialized events

func GetEventsFromFileSortedByTimestamp(serializedDecoders []func(sEvt []byte) (stdtypes.Event, error), filenames ...string) []stdtypes.Event {
	recordingLog := getRecordingLog(filenames...)

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
		}
	}

	return eventLog
}

func GetEventsFromFileSortedByVectorClock(serializedDecoders []func(sEvt []byte) (stdtypes.Event, error), filenames ...string) []stdtypes.Event {
	recordingLog := getRecordingLog(filenames...)
	type vcEvent struct {
		vc    *vectorclock.VectorClock
		event stdtypes.Event
	}

	vcEvents := make([]vcEvent, 0, len(recordingLog))
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
			vcEvents = append(vcEvents, vcEvent{getVC(event), event})
		}
	}

	vcEvents = sliceutil.TopologicalSortFunc(vcEvents, func(a, b vcEvent) bool {
		return vectorclock.Less(a.vc, b.vc)
	})

	return sliceutil.Transform(vcEvents, func(_ int, vcE vcEvent) stdtypes.Event { return vcE.event })
}

func getVC(e stdtypes.Event) *vectorclock.VectorClock {
	v, err := e.GetMetadata("vc")
	if err != nil {
		panic(fmt.Sprintf("error: %v, event: %v", err, e))
	}
	vc, ok := v.(map[string]interface{})
	if !ok {
		panic("fuck")
	}
	vcc, ok := vc["V"].(map[string]interface{})
	if !ok {
		panic("fuckfuck")
	}
	vcR := make(map[stdtypes.NodeID]uint32)
	for k, v := range vcc {
		vcR[stdtypes.NodeID(k)] = uint32(v.(float64))
	}

	return vectorclock.VectorClockFromMap(vcR)
}

func getRecordingLog(filenames ...string) []*recordingpb.Entry {
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

	return recordingLog
}
