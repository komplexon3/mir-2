package nodeinstance

import (
	"context"
	"sync"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/logging"
)

type NodesRunner struct {
	nodeInstances NodeInstances
	doneC         chan struct{}
	isDoneC       chan struct{}
}

func NewNodesRunner(nodeInstances NodeInstances) *NodesRunner {
	return &NodesRunner{
		nodeInstances: nodeInstances,
		doneC:         make(chan struct{}),
		isDoneC:       make(chan struct{}),
	}
}

// NOTE: Can only be stopped by cancelling the context, maybe doneC pattern would be nice
// but adding the scaffolding for this is probably not worth it
func (r *NodesRunner) Run(
	ctx context.Context,
	logger logging.Logger,
) error {
	defer close(r.isDoneC)
	wg := &sync.WaitGroup{}

	var err error
	errChan := make(chan error)

	for nodeID, nodeInstance := range r.nodeInstances {
		wg.Add(1)
		go func() {
			defer wg.Done()
			nodeInstance.Setup()
			select {
			case errChan <- nodeInstance.Run(ctx):
			default:
			}
			logger.Log(logging.LevelDebug, "node has stopped", "node", nodeID)
		}()
	}

ShutdownLoop:
	for {
		select {
		case nodeErr := <-errChan:
			if nodeErr != nil && nodeErr != mir.ErrStopped {
				// node failed, kill all other nodes
				err = nodeErr
				logger.Log(logging.LevelDebug, "node failed with error", "error", nodeErr)
				break ShutdownLoop
			}
		case <-r.doneC:
			break ShutdownLoop
		case <-ctx.Done():
			break ShutdownLoop
		}
	}

	for _, nodeInstance := range r.nodeInstances {
		wg.Add(1)
		go func() {
			defer wg.Done()
			nodeInstance.Stop()
			nodeInstance.Cleanup()
		}()
	}

	wg.Wait()

	return err
}

func (r *NodesRunner) Stop() error {
	close(r.doneC)
	<-r.isDoneC
	return nil
}