package adversary

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

// need
//

type Adversary struct {
	nodeInstances               map[stdtypes.NodeID]NodeInstance
	byzantineCortexCreeperLinks map[stdtypes.NodeID]CortexCreeperLink
}

// Setup parameters
type params struct {
	nodeIds        []stdtypes.NodeID
	byzantineNodes []stdtypes.NodeID
}

type nodeConfig struct {
	eventsIn  chan stdtypes.EventList
	eventsOut chan stdtypes.EventList
}

func NewAdversary[T interface{}](
	createNodeInstance NodeInstanceCreationFunc[T],
	nodeConfigs NodeConfigs[T],
	byzantineNodes []stdtypes.NodeID,
) (*Adversary, error) {
	if len(nodeConfigs) < len(byzantineNodes) {
		return nil, es.Errorf("number of byzantine cannot be larger than number of node configs")
	}
	baseNodeLogger := logging.ConsoleWarnLogger

	nodeInstances := make(map[stdtypes.NodeID]NodeInstance)
	byzantineCortexCreeperLinks := make(map[stdtypes.NodeID]CortexCreeperLink)

	for nodeID, config := range nodeConfigs {
		nodeLogger := logging.Decorate(baseNodeLogger, "Node %s - ", nodeID)
		var cortexCreeper CortexCreeper
		if slices.Contains(byzantineNodes, nodeID) {
			cortexCreeperLink := NewCortexCreeperLink()
			byzantineCortexCreeperLinks[nodeID] = cortexCreeperLink
			cortexCreeper = NewByzantineCortexCreeper(cortexCreeperLink)
		} else {
			cortexCreeper = NewBenignCortexCreeper()
		}
		nodeInstance, err := createNodeInstance(nodeID, config, cortexCreeper, nodeLogger)
		if err != nil {
			return nil, es.Errorf("Failed to create node instance with id %s: %v", nodeID, err)
		}
		nodeInstances[nodeID] = nodeInstance
	}

	return &Adversary{nodeInstances, byzantineCortexCreeperLinks}, nil
}

func (a *Adversary) RunPlugin(plugin *Plugin, c context.Context) error {
	// setup all nodes and run them
	wg := sync.WaitGroup{}
	runnerContext, cancelRunnerContext := context.WithCancel(c)
	defer cancelRunnerContext()
	for nodeId, nodeInstance := range a.nodeInstances {
		errChan := make(chan error)

		go func(instance NodeInstance) {
			defer close(errChan)
			wg.Add(1)
			instance.Setup()
			errChan <- instance.GetNode().Run(context.Background())
		}(nodeInstance)

		go func(instance NodeInstance, id stdtypes.NodeID, ctx context.Context, cancel context.CancelFunc) {
			defer wg.Done()
			defer instance.Cleanup()
			select {
			case <-ctx.Done():
				instance.GetNode().Stop()
			case err := <-errChan:
				if err != mir.ErrStopped {
					// node failed, kill all other nodes
					fmt.Printf("Node %s failed with error: %v", id, err)
				}
				cancel()
			}
		}(nodeInstance, nodeId, runnerContext, cancelRunnerContext)
	}

	// do some other stuff
	// run plugin for every byz node
	for nodeId, cortexCreeperLink := range a.byzantineCortexCreeperLinks {
		go func(id stdtypes.NodeID, ccl CortexCreeperLink, ctx context.Context, cancel context.CancelFunc) {
			wg.Add(1)
			defer wg.Done()

			for {
				select {
				case evts := <-ccl.toAdversary:
					evtsWithNodeIdMeta := addNodeIdToEvents(*evts, id)
					newEvts, err := plugin.processEvents(&evtsWithNodeIdMeta)
					if err != nil {
						fmt.Printf("Plugin %s failed to apply events: %v", plugin.name, err)
						cancel()
					}
					ccl.toNode <- newEvts
				case <-ctx.Done():
					break
				}
			}
		}(nodeId, cortexCreeperLink, runnerContext, cancelRunnerContext)
	}

	// if we are done but the nodes are still running...
	<-c.Done()
	// waiting for all contexts to terminate
	wg.Wait()

	return nil
}

func (a *Adversary) RunExperiment(plugin *Plugin, puppeteer Puppeteer) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go a.RunPlugin(plugin, ctx)

	err := puppeteer.Run(maputil.GetValues(a.nodeInstances))
	if err != nil {
		return err
	}
	return nil
}

func addNodeIdToEvents(events stdtypes.EventList, nodeId stdtypes.NodeID) stdtypes.EventList {
	return *stdtypes.ListOf(sliceutil.Transform(events.Slice(), func(_ int, event stdtypes.Event) stdtypes.Event {
		newEvent, err := event.SetMetadata("nodeId", nodeId)
		if err != nil {
			// TODO handle errors nicely
			panic(es.Errorf("failed to set node id in metadata: %v", err))
		}
		return newEvent
	})...)
}
