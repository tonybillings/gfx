package gfx

/******************************************************************************
 Async Wrappers
******************************************************************************/

type asyncBoolInvocation struct {
	Func       func() bool
	ReturnChan chan bool
}

type asyncVoidInvocation struct {
	Func     func()
	DoneChan chan bool
}

/******************************************************************************
 New Functions
******************************************************************************/

func newAsyncBoolInvocation(f func() bool) *asyncBoolInvocation {
	return &asyncBoolInvocation{
		Func:       f,
		ReturnChan: make(chan bool, 1),
	}
}

func newAsyncVoidInvocation(f func()) *asyncVoidInvocation {
	return &asyncVoidInvocation{
		Func:     f,
		DoneChan: make(chan bool, 1),
	}
}
