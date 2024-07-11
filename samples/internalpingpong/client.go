package main

import (
	"context"
	"fmt"
	"os"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/stdtypes"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/samples/internalpingpong/pingpong"
)

const numModules = 10

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {

	var logger logging.Logger
	logger = logging.ConsoleTraceLogger // Print trace-level info.

	modules := make(modules.Modules, numModules)
	for i := range numModules {
		mID := stdtypes.ModuleID(fmt.Sprint(i))
		otherModuleIDs := make([]stdtypes.ModuleID, 0, numModules-1)
		for j := range numModules {
			if i == j {
				continue
			}
			otherModuleIDs = append(otherModuleIDs, stdtypes.ModuleID(fmt.Sprint(j)))
		}
		modules[mID] = pingpong.NewModule(pingpong.ModuleConfig{
			Self:   mID,
			Others: otherModuleIDs,
		}, logger)
	}

	// create a Mir node
	node, err := mir.NewNode("internalpingpongtest", mir.DefaultNodeConfig().WithLogger(logger), modules, nil)
	if err != nil {
		return es.Errorf("error creating a Mir node: %w", err)
	}

	// run the node
	err = node.Run(context.Background())
	if err != nil {
		return es.Errorf("error running node: %w", err)
	}

	return nil
}
