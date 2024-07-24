package nodeinstance

import (
	"context"
	"sync"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/logging"
)

// NOTE: Can only be stopped by cancelling the context, maybe doneC pattern would be nice
// but adding the scaffolding for this is probably not worth it
func RunNodes(
	ctx context.Context,
	nodeInstances NodeInstances,
	logger logging.Logger,
) error {
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for nodeID, nodeInstance := range nodeInstances {
		wg.Add(1)
		go func() {
			errChan := make(chan error)
			defer wg.Done()
			defer nodeInstance.Stop()
			defer nodeInstance.Cleanup()
			nodeInstance.Setup()

			go func() {
				errChan <- nodeInstance.Run(ctx)
				logger.Log(logging.LevelDebug, "node has stopped", "node", nodeID)
			}()

			select {
			case err := <-errChan:
				if err != nil && err != mir.ErrStopped {
					// node failed, kill all other nodes
					logger.Log(logging.LevelDebug, "node failed with error", "node", nodeID, "error", err)
				}
			case <-ctx.Done():
			}
		}()
	}

	<-ctx.Done()

	wg.Wait()

	return nil
}
