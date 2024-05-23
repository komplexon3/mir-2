package adversary

import "github.com/filecoin-project/mir/stdtypes"

type Puppeteer interface {
	Run(nodeInstances map[stdtypes.NodeID]NodeInstance) error // runs the puppeteer
}
