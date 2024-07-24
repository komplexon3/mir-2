package vcinterceptor

import (
	"github.com/filecoin-project/mir/fuzzer2/interceptors/vcinterceptor/vectorclock"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

// TODO: instrument each event with current (incremented) vc
// for every incoming message, unwrap and combine with current vc (and increment?)
// for outgoing message, wrap and add current vc (incremented?)

const vcKey = "vc"

// NOTE: this is just an interceptor so no concurrency - why did I bother with mutexes?

type VectorClockInterceptor struct {
	vc     vectorclock.VectorClock
	nodeID stdtypes.NodeID
}

func New(nodeId stdtypes.NodeID) *VectorClockInterceptor {
	return &VectorClockInterceptor{
		nodeID: nodeId,
		vc:     *vectorclock.NewVectorClock(),
	}
}

func (vci *VectorClockInterceptor) Intercept(events *stdtypes.EventList) (*stdtypes.EventList, error) {
	newEvents := stdtypes.EmptyList()
	iter := events.Iterator()
	for ev := iter.Next(); ev != nil; ev = iter.Next() {
		var nEv stdtypes.Event
		var err error

		// TODO: proto msgs?
		switch evT := ev.(type) {
		case *stdevents.MessageReceived:
			vcMsgRaw, err := evT.GetMetadata("vc")
			if err != nil {
				return nil, err
			}
			vcMsg, err := RawVcIntoVC(vcMsgRaw)
			if err != nil {
				panic("couldn't convert raw metadata into vc")
			}
			vci.vc.CombineAndIncrement(vcMsg, vci.nodeID)
			vc := vci.vc.Clone()
			nEv, err = ev.SetMetadata(vcKey, vc)
			if err != nil {
				return nil, err
			}
		case *eventpb.Event:
			switch pbEvT := evT.Type.(type) {
			case *eventpb.Event_Transport:
				return nil, es.Errorf("unexpected proto message event, all proto transport events should be wrapped to be processed by vcinterceptor : %T", pbEvT)
			default:
				vci.vc.Increment(vci.nodeID)
				vc := vci.vc.Clone()
				nEv, err = ev.SetMetadata(vcKey, vc)
				if err != nil {
					return nil, err
				}
			}
		default:
			vci.vc.Increment(vci.nodeID)
			vc := vci.vc.Clone()
			nEv, err = ev.SetMetadata(vcKey, vc)
			if err != nil {
				return nil, err
			}
		}

		newEvents.PushBack(nEv)
	}
	return newEvents, nil
}
