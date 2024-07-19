package actions

import (
	wr "github.com/mroth/weightedrand/v2"
)

type RandomActions struct {
	chooser *wr.Chooser[Action, int]
}

type WeightedAction struct {
	action Action
	weight int
}

func NewWeightedAction(action Action, weight int) WeightedAction {
	return WeightedAction{
		action: action,
		weight: weight,
	}
}

func NewRandomActions(weightedActions []WeightedAction) (*RandomActions, error) {
	choices := make([]wr.Choice[Action, int], len(weightedActions))
	for i, wa := range weightedActions {
		choices[i] = wr.NewChoice(wa.action, wa.weight)
	}

	chooser, err := wr.NewChooser(choices...)
	if err != nil {
		return nil, err
	}

	return &RandomActions{
		chooser,
	}, nil
}

func (ra *RandomActions) SelectAction() Action {
	return ra.chooser.Pick()
}
