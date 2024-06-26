package centraladversay

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/vcinterceptor"
	"github.com/filecoin-project/mir/pkg/vcinterceptor/vectorclock"
	"github.com/filecoin-project/mir/stdtypes"
)

// TODOs:
// - Make it an actual priority queu
// - deal with the vc mess

type Aggregator struct {
	pq               *priorityQueue
	eventSelectCases []reflect.SelectCase
	subscribed       bool
}

type priorityQueue struct {
	queue          []stdtypes.Event
	maxLen         int
	spaceAvailable chan struct{}
	dataAvailable  chan struct{}
	mutex          *sync.Mutex
}

func (pq *priorityQueue) swap(i, j int) {
	pq.queue[i], pq.queue[j] = pq.queue[j], pq.queue[i]
}

// Blocks if no space is available
func (pq *priorityQueue) PushEvent(e stdtypes.Event, ctx context.Context) {
	// Add the new element to the end of the heap
	pq.mutex.Lock()
	fmt.Printf("Pushing event, current len %d\n", len(pq.queue))

	// block until space is available
	for len(pq.queue) == pq.maxLen {
		fmt.Println("Full - can't push")
		pq.mutex.Unlock()
		fmt.Println("Blocking for full")
		select {
		case <-ctx.Done():
			return
		case <-pq.spaceAvailable:
			fmt.Println("Recv space available")
			pq.mutex.Lock()
			break
		}
	}

	defer pq.mutex.Unlock()

	if len(pq.queue) == 0 {
		// indicate that data is available again
		defer (func() {
			select {
			case pq.dataAvailable <- struct{}{}:
				fmt.Println("data available notification")
			default:
				fmt.Println("NO data available notification")
			}
		})()
	}

	// TODO: probably makes sense to store vc in structure directly instead of accessing it all the time

	pq.queue = append(pq.queue, e)
	var vcindex, vcmid *vectorclock.VectorClock
	index := len(pq.queue) - 1
	mid := (index - 1) / 2

	switch eT := pq.queue[index].(type) {
	case *eventpb.Event:
		vcindex, _ = vcinterceptor.ExtractVCFromPbEvent(eT)
	default:
		vcindexR, _ := pq.queue[index].GetMetadata("vc")
		vcindex = vcindexR.(*vectorclock.VectorClock)
	}

	switch eT := pq.queue[mid].(type) {
	case *eventpb.Event:
		vcmid, _ = vcinterceptor.ExtractVCFromPbEvent(eT)
	default:
		vcmidR, _ := pq.queue[mid].GetMetadata("vc")
		vcmid = vcmidR.(*vectorclock.VectorClock)
	}

	for index > 0 &&
		vectorclock.Less(vcindex, vcmid) {
		pq.swap(index, mid)

		vcindex = vcmid
		index = mid
		mid = (index - 1) / 2
		switch eT := pq.queue[mid].(type) {
		case *eventpb.Event:
			vcmid, _ = vcinterceptor.ExtractVCFromPbEvent(eT)
		default:
			vcmidR, _ := pq.queue[mid].GetMetadata("vc")
			vcmid = vcmidR.(*vectorclock.VectorClock)
		}
	}

	fmt.Printf("Pushed event, current len %d\nContent:", len(pq.queue))
	for _, ev := range pq.queue {
		fmt.Printf("%v, ", ev)
	}
	fmt.Println("")
}

func (pq *priorityQueue) PopEvent(ctx context.Context) stdtypes.Event {
	pq.mutex.Lock()
	fmt.Printf("Popping event, current len %d\n", len(pq.queue))

	for len(pq.queue) == 0 {
		fmt.Println("Empty - can't pop")
		pq.mutex.Unlock()
		fmt.Println("Blocking for empty")
		select {
		case <-ctx.Done():
			return nil
		case <-pq.dataAvailable:
			fmt.Println("Recv data available")
			pq.mutex.Lock()
			break
		}
	}

	defer pq.mutex.Unlock()

	if len(pq.queue) == pq.maxLen {
		// indicate that there is space available
		defer (func() {
			select {
			case pq.spaceAvailable <- struct{}{}:
				fmt.Println("space available notification")
			default:
				fmt.Println("NO space available notification")
			}
		})()
	}

	if len(pq.queue) <= 0 {
		return nil
	}
	item := pq.queue[len(pq.queue)-1]
	pq.queue = pq.queue[0 : len(pq.queue)-1 : pq.maxLen]

	fmt.Printf("Popped event, current len %d\nContent:", len(pq.queue))
	for _, ev := range pq.queue {
		fmt.Printf("%v, ", ev)
	}
	fmt.Println("")
	return item
}

func NewAggregator(maxQueue int, eventChans ...<-chan *stdtypes.EventList) *Aggregator {
	selectCases := make([]reflect.SelectCase, len(eventChans))
	for i, ch := range eventChans {
		selectCases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}
	return &Aggregator{
		pq: &priorityQueue{
			queue:          make([]stdtypes.Event, 0, maxQueue),
			maxLen:         maxQueue,
			mutex:          &sync.Mutex{},
			spaceAvailable: make(chan struct{}),
			dataAvailable:  make(chan struct{}),
		},
		eventSelectCases: selectCases,
		subscribed:       false,
	}
}

func (a *Aggregator) Run(ctx context.Context) {
	closedChans := 0
	for closedChans < len(a.eventSelectCases) {
		_, val, ok := reflect.Select(a.eventSelectCases)
		if !ok {
			closedChans++
			continue
		}

		el := val.Interface().(*stdtypes.EventList)
		elIter := el.Iterator()
		for event := elIter.Next(); event != nil; event = elIter.Next() {
			select {
			case <-ctx.Done():
				return
			default:
				a.pq.PushEvent(event, ctx)
			}
		}
	}
}

func (a *Aggregator) Subscribe(ctx context.Context) <-chan stdtypes.Event {
	if a.subscribed {
		// ugly but will do for now
		panic("aggregator nly permits one subscriber")
	}
	outChan := make(chan stdtypes.Event)

	go func() {
		defer close(outChan)
		for {
			event := a.pq.PopEvent(ctx)
			outChan <- event
		}
	}()

	return outChan
}
