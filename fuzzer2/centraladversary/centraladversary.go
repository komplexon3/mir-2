package centraladversay

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/fuzzer2/actions"
	"github.com/filecoin-project/mir/fuzzer2/centraladversary/cortexcreeper"
	"github.com/filecoin-project/mir/fuzzer2/heartbeat"
	"github.com/filecoin-project/mir/fuzzer2/nodeinstance"
	"github.com/filecoin-project/mir/fuzzer2/puppeteer"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"

	es "github.com/go-errors/errors"
)

type Adversary struct {
	nodeInstances  map[stdtypes.NodeID]nodeinstance.NodeInstance
	cortexCreepers []*cortexcreeper.CortexCreeper
	actionSelector actions.Actions
}

func NewAdversary[T interface{}](
	createNodeInstance nodeinstance.NodeInstanceCreationFunc[T],
	nodeConfigs nodeinstance.NodeConfigs[T],
	weightedActions []actions.WeightedAction,
	logger logging.Logger,
) (*Adversary, error) {

	nodeInstances := make(map[stdtypes.NodeID]nodeinstance.NodeInstance)
	cortexCreepers := make([]*cortexcreeper.CortexCreeper, 0, len(nodeConfigs))

	for nodeID, config := range nodeConfigs {
		nodeLogger := logging.Decorate(logger, string(nodeID)+" ")
		// nodeLogger := logger
		cortexCreeper := cortexcreeper.NewCortexCreeper()
		cortexCreepers = append(cortexCreepers, cortexCreeper)
		nodeInstance, err := createNodeInstance(nodeID, config, cortexCreeper, nodeLogger)
		if err != nil {
			return nil, es.Errorf("Failed to create node instance with id %s: %v", nodeID, err)
		}
		nodeInstances[nodeID] = nodeInstance
	}

	actionSelector, err := actions.NewRandomActions(weightedActions)
	if err != nil {
		return nil, err
	}

	return &Adversary{nodeInstances, cortexCreepers, actionSelector}, nil
}

func (a *Adversary) RunNodes(ctx context.Context) error {
	// setup all nodes and run them
	wg := &sync.WaitGroup{}
	nodesContext, advCancel := context.WithCancel(ctx)
	defer advCancel()
	for nodeId, nodeInstance := range a.nodeInstances {
		go func() {
			errChan := make(chan error)
			wg.Add(1)
			defer wg.Done()
			defer nodeInstance.Stop()
			defer nodeInstance.Cleanup()
			nodeInstance.Setup()
			go func() {
				errChan <- nodeInstance.Run(nodesContext)
				fmt.Printf("node %s stopped\n", nodeId)
			}()

			select {
			case err := <-errChan:
				if err != nil && err != mir.ErrStopped {
					// node failed, kill all other nodes
					fmt.Printf("Node %s failed with error: %v", nodeId, err)
				}
			case <-nodesContext.Done():

			}
			fmt.Printf("node %s done\n", nodeId)
		}()
	}

	time.Sleep(time.Second)

	for _, nodeInstance := range a.nodeInstances {
		// inject heartbeat start
		nodeInstance.GetNode().InjectEvents(nodesContext,
			stdtypes.ListOf(
				stdevents.NewTimerRepeat("timer",
					time.Second,
					stdtypes.RetentionIndex(^uint64(0)),
					heartbeat.NewHeartbeat("null"),
				),
			),
		)
	}
	<-nodesContext.Done()
	fmt.Println("node shutdown received")

	wg.Wait()
	fmt.Println("node wg done")

	return nil
}

func (a *Adversary) RunCentralAdversary(maxEvents, maxHearbeatsInactive int, ctx context.Context) {
	fmt.Println("Running CA")
	eventCount := 0
	heartbeatCount := 0

	for {
		selectCases := make([]reflect.SelectCase, 0, len(a.cortexCreepers)+1)
		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		})
		for _, cc := range a.cortexCreepers {
			selectCases = append(selectCases, reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(cc.GetEventsIn()),
			})
		}

		ind, value, _ := reflect.Select(selectCases)
		if ind == 0 {
			// context was cancelled
			return
		}

		// process this event
		elIterator := value.Interface().(*stdtypes.EventList).Iterator()
		for event := elIterator.Next(); event != nil; event = elIterator.Next() {
			eventCount++
			switch event.(type) {
			case *heartbeat.Heartbeat:
				heartbeatCount++
			default:
				heartbeatCount = 0
			}

			fmt.Printf("#e: %d, #h: %d\n", eventCount, heartbeatCount)

			if heartbeatCount > maxHearbeatsInactive || eventCount > maxEvents {
				fmt.Println("wrapping up...")
				return
			}

			fmt.Println("+++ Processing event +++")
			// TODO: do stuff -> right now we are simply injecting the events
			a.cortexCreepers[ind-1].PushEvents(stdtypes.ListOf(event))
		}

	}
}

func (a *Adversary) RunExperiment(puppeteer puppeteer.Puppeteer, maxEvents, maxHeartbeatsInactive int) error {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	defer cancel()

	nodesErr := make(chan error)
	go func() {
		wg.Add(1)
		defer wg.Done()
		nodesErr <- a.RunNodes(ctx)
		fmt.Println("Run nodes done")
	}()

	go func() {
		wg.Add(1)
		defer wg.Done()
		defer cancel() // cancel to stop nodes when adv had enough
		a.RunCentralAdversary(maxEvents, maxHeartbeatsInactive, ctx)
		fmt.Println("CA done")
	}()

	err := puppeteer.Run(a.nodeInstances)
	fmt.Println("Ran puppeteer")
	if err != nil {
		return err
	}

	err = <-nodesErr
	if err != nil {
		return err
	}

	wg.Wait()
	fmt.Println("done?")

	return nil
}
