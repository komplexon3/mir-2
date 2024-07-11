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

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {

	var logger logging.Logger
	logger = logging.ConsoleTraceLogger // Print trace-level info.

	aModule := pingpong.NewModule(pingpong.ModuleConfig{
		Self:  stdtypes.ModuleID("A"),
		Other: stdtypes.ModuleID("B"),
	}, logger)
	bModule := pingpong.NewModule(pingpong.ModuleConfig{
		Self:  stdtypes.ModuleID("B"),
		Other: stdtypes.ModuleID("A"),
	}, logger)

	m := map[stdtypes.ModuleID]modules.Module{
		"A": aModule,
		"B": bModule,
	}

	// create a Mir node
	node, err := mir.NewNode("internalpingpongtest", mir.DefaultNodeConfig().WithLogger(logger), m, nil)
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
