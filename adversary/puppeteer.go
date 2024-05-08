package adversary

type Puppeteer interface {
  Run(nodeInstances []NodeInstance) error// runs the puppeteer
}
