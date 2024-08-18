package interceptors

import (
	"github.com/filecoin-project/mir/fuzzer/cortexcreeper"
	"github.com/filecoin-project/mir/fuzzer/interceptors/deepcopymsgs"
	msgmetadata "github.com/filecoin-project/mir/fuzzer/interceptors/msgMetadata"
	"github.com/filecoin-project/mir/fuzzer/interceptors/nomulticast"
	"github.com/filecoin-project/mir/fuzzer/interceptors/vcinterceptor"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/stdtypes"
)

type FuzzerInterceptor struct {
	preCAInterceptor  eventlog.Interceptor
	postCAInterceptor eventlog.Interceptor
}

// type nilSnitch struct {
// 	prevInterceptor string
// }
//
// func (s *nilSnitch) Intercept(events *stdtypes.EventList) (*stdtypes.EventList, error) {
// 	if events == nil {
// 		panic(fmt.Sprint(s.prevInterceptor, "returned nil"))
// 	}
//
// 	return events, nil
// }

// Primary purpose: simplify the logs a bit bc it is becomming quite hard to sift
// through them if all events show up in there multiple times

func NewFuzzerInterceptor(nodeID stdtypes.NodeID, cortexCreeper *cortexcreeper.CortexCreeper, logger logging.Logger, extraInterceptors ...eventlog.Interceptor) (*FuzzerInterceptor, error) {
	msgMetadataInterceptorIn, msgMetadataInterceptorOut := msgmetadata.NewMsgMetadataInterceptorPair(logger, "vc", "msgID")

	postCAInterceptors := []eventlog.Interceptor{
		vcinterceptor.New(nodeID),
		msgMetadataInterceptorOut,
	}
	postCAInterceptors = append(postCAInterceptors, extraInterceptors...)

	return &FuzzerInterceptor{
		preCAInterceptor: eventlog.MultiInterceptor(
			msgMetadataInterceptorIn,
			&nomulticast.NoMulticast{},
			&deepcopymsgs.DeepCopyMsgs{},
			cortexCreeper,
		),
		postCAInterceptor: eventlog.MultiInterceptor(
			postCAInterceptors...,
		),
	}, nil
}

func (i *FuzzerInterceptor) Intercept(events *stdtypes.EventList) (*stdtypes.EventList, error) {
	// if injected, pass through second half, otherwise, pass through first half
	// if one element is injected then all in the list are so we must only check one
	if events.Len() >= 1 {
		if _, err := events.Slice()[0].GetMetadata("injected"); err == nil {
			return i.postCAInterceptor.Intercept(events)
		}
	}

	return i.preCAInterceptor.Intercept(events)
}
