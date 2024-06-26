package centraladversay

import (
	"context"
	"fmt"
	"sync"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/fuzzer/actions"
	"github.com/filecoin-project/mir/fuzzer/cortexcreeper"
	"github.com/filecoin-project/mir/fuzzer/nodeinstance"
	"github.com/filecoin-project/mir/fuzzer/puppeteer"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
	"github.com/filecoin-project/mir/stdtypes"
	"github.com/gogo/protobuf/test/stdtypes"

	es "github.com/go-errors/errors"
)

const MAX_QUEUE_SIZE = 5

type Adversary struct {
	nodeInstances               map[stdtypes.NodeID]nodeinstance.NodeInstance
	byzantineCortexCreeperLinks map[stdtypes.NodeID]cortexcreeper.CortexCreeperLink
	actionSelector              actions.Actions
	eventAggregator             *Aggregator
}

type nodeConfig struct {
	eventsIn  chan stdtypes.EventList
	eventsOut chan stdtypes.EventList
}

func NewAdversary[T interface{}](
	createNodeInstance nodeinstance.NodeInstanceCreationFunc[T],
	nodeConfigs nodeinstance.NodeConfigs[T],
	weightedActions []actions.WeightedAction,
	logger logging.Logger,
) (*Adversary, error) {

	nodeInstances := make(map[stdtypes.NodeID]nodeinstance.NodeInstance)
	byzantineCortexCreeperLinks := make(map[stdtypes.NodeID]cortexcreeper.CortexCreeperLink)

	for nodeID, config := range nodeConfigs {
		nodeLogger := logging.Decorate(logger, string(nodeID)+" ")
		// nodeLogger := logger
		cortexCreeperLink := cortexcreeper.NewCortexCreeperLink()
		byzantineCortexCreeperLinks[nodeID] = cortexCreeperLink
		cortexCreeper := cortexcreeper.NewByzantineCortexCreeper(cortexCreeperLink)
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

	aggregator := NewAggregator(
		MAX_QUEUE_SIZE,
		sliceutil.Transform(
			maputil.GetValues(byzantineCortexCreeperLinks),
			func(_ int, ccl cortexcreeper.CortexCreeperLink) <-chan *stdtypes.EventList {
				return ccl.ToAdversary
			},
		)...,
	)

	return &Adversary{nodeInstances, byzantineCortexCreeperLinks, actionSelector, aggregator}, nil
}

func (a *Adversary) Run(c context.Context) error {
	// setup all nodes and run them
	wg := sync.WaitGroup{}
	runnerContext, cancelRunnerContext := context.WithCancel(c)
	defer cancelRunnerContext()
	for nodeId, nodeInstance := range a.nodeInstances {
		errChan := make(chan error)
		ctx := context.Background()

		go func(instance nodeinstance.NodeInstance) {
			defer close(errChan)
			wg.Add(1)
			instance.Setup()
			errChan <- instance.Run(ctx)
		}(nodeInstance)

		go func(instance nodeinstance.NodeInstance, id stdtypes.NodeID, ctx context.Context, cancel context.CancelFunc) {
			defer wg.Done()
			defer instance.Cleanup()
			select {
			case <-ctx.Done():
				instance.Setup()
				instance.Stop()
			case err := <-errChan:
				if err != mir.ErrStopped {
					// node failed, kill all other nodes
					fmt.Printf("Node %s failed with error: %v", id, err)
				}
				cancel()
			}
		}(nodeInstance, nodeId, runnerContext, cancelRunnerContext)
	}

	// should probably be added to waitgroup
	go a.eventAggregator.Run(c)
	// blocking
	a.RunCentralAdversary(c)

	// if we are done but the nodes are still running...
	// props not needed anymore
	<-c.Done()
	// waiting for all contexts to terminate
	wg.Wait()

	return nil
}

func (a *Adversary) RunCentralAdversary(ctx context.Context) {
	fmt.Println("Running CA")
	for event := range a.eventAggregator.Subscribe(ctx) {
		if event == nil {
			return
		}
		originNodeIdR, _ := event.GetMetadata("node")
		var originNodeId stdtypes.NodeID

		// TODO: This code looks like I had a stroke while writing it - there MUST be a better way to do this
		switch onidt := originNodeIdR.(type) {
		case string:
			originNodeId = stdtypes.NodeID(onidt)
		case stdtypes.NodeID:
			originNodeId = onidt
		default:
			originNodeId = stdtypes.NodeID(onidt.(string))
		}

		fmt.Printf("+++ Processing event %v +++\n", event)
		// do stuff
		a.byzantineCortexCreeperLinks[stdtypes.NodeID(originNodeId.(string))].ToNode <- stdtypes.ListOf(event)
		fmt.Printf("+++ Processed event %v +++\n", event)
	}
}

func (a *Adversary) RunExperiment(puppeteer puppeteer.Puppeteer) error {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	defer cancel()
	go func() {
		wg.Add(1)
		defer wg.Done()
		a.Run(ctx)
	}()

	err := puppeteer.Run(a.nodeInstances)
	fmt.Println("Ran puppeteer")
	if err != nil {
		return err
	}

	wg.Wait()

	return nil
}
