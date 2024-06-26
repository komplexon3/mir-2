package puppeteer

import (
	"github.com/filecoin-project/mir/fuzzer2/nodeinstance"
	"github.com/filecoin-project/mir/stdtypes"
)

type Puppeteer interface {
	Run(nodeInstances map[stdtypes.NodeID]nodeinstance.NodeInstance) error // runs the puppeteer
}