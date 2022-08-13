package action

type WrapAction struct {
	innerAction Action
	aroundFn    AroundFunc
}

type AroundFunc func(ctx Context, innerAction Action) (Result, error)

func Wrap(action Action, aroundFn AroundFunc) *WrapAction {
	return &WrapAction{action, aroundFn}
}

func (a *WrapAction) Run(ctx Context) (Result, error) {
	return a.aroundFn(ctx, a.innerAction)
}
