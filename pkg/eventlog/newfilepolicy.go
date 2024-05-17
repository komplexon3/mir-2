package eventlog

import (
	"github.com/filecoin-project/mir/pkg/pb/apppb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/stdtypes"
)

// EventNewEpochLogger returns a file that splits an event list into multiple lists
// every time an event eventpb.Event_NewLogFile is found
func EventNewEpochLogger(appModuleID stdtypes.ModuleID) func(*stdtypes.EventList) []*stdtypes.EventList {
	eventNewLogFileLogger := func(event *stdtypes.Event) bool {
		pbevent, ok := (*event).(*eventpb.Event)
		if !ok {
			return false
		}

		appEvent, ok := pbevent.Type.(*eventpb.Event_App)
		if !ok {
			return false
		}

		_, ok = appEvent.App.Type.(*apppb.Event_NewEpoch)
		return ok && pbevent.Dest() == appModuleID
	}
	return EventTrackerLogger(eventNewLogFileLogger)
}

// EventTrackerLogger returns a function that tracks every single event of EventList and
// creates a new file for every event such that newFile(event) = True
func EventTrackerLogger(newFile func(event *stdtypes.Event) bool) func(*stdtypes.EventList) []*stdtypes.EventList {
	return func(evts *stdtypes.EventList) []*stdtypes.EventList {
		var result []*stdtypes.EventList
		currentChunk := stdtypes.EmptyList()

		for _, event := range evts.Slice() {
			if newFile(&event) {
				result = append(result, currentChunk)
				currentChunk = stdtypes.ListOf(event)
			} else {
				currentChunk.PushBack(event)
			}
		}

		// If there is a remaining chunk with fewer than the desired number of events, append it to the result
		if currentChunk.Len() > 0 {
			result = append(result, currentChunk)
		}

		return result
	}
}

// EventLimitLogger returns a function for the interceptor that splits the logging file
// every eventLimit number of events
func EventLimitLogger(eventLimit int64) func(*stdtypes.EventList) []*stdtypes.EventList {
	var eventCount int64
	return func(evts *stdtypes.EventList) []*stdtypes.EventList {
		var result []*stdtypes.EventList
		currentChunk := stdtypes.EmptyList()

		// Iterate over the events in the input slice
		for _, event := range evts.Slice() {
			// Add the current element to the current chunk
			currentChunk.PushBack(event)
			eventCount++
			// If the current chunk has the desired number of events, append it to the result and start a new chunk
			if eventCount%eventLimit == 0 {
				result = append(result, currentChunk)
				currentChunk = stdtypes.EmptyList()
			}
		}

		// If there is a remaining chunk with fewer than the desired number of events, append it to the result
		if currentChunk.Len() > 0 {
			result = append(result, currentChunk)
		}

		return result
	}
}

func OneFileLogger() func(*stdtypes.EventList) []*stdtypes.EventList {
	return func(evts *stdtypes.EventList) []*stdtypes.EventList {
		return []*stdtypes.EventList{evts}
	}
}
