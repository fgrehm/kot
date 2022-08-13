package action

import (
	"fmt"
	"strings"
)

type CompositeAction struct {
	allowErrors bool
	actions     []Action
}

func Composite(actions ...Action) *CompositeAction {
	return &CompositeAction{actions: actions}
}

func (a *CompositeAction) Run(ctx Context) (Result, error) {
	allErrors := []string{}
	result := Result{}

	for _, action := range a.actions {
		actionRes, err := action.Run(ctx)
		result = result.Merge(actionRes)
		if err != nil {
			if !a.allowErrors {
				return result, err
			}
			allErrors = append(allErrors, err.Error())
		}
		if result.Halt {
			break
		}
	}

	result.Halt = false // "Consume" the halt, if any
	if len(allErrors) == 0 {
		return result, nil
	}

	return result, fmt.Errorf(`one or more errors occurred: ["%s"]`, strings.Join(allErrors, `", "`))
}

func (a *CompositeAction) AllowErrors() *CompositeAction {
	a.allowErrors = true
	return a
}
