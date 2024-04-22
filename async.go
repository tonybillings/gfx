package gfx

/******************************************************************************
 Async Wrappers
******************************************************************************/

type asyncVoidInvocation struct {
	Func     func()
	DoneChan chan bool
}

type asyncBoolInvocation struct {
	Func       func() bool
	ReturnChan chan bool
}

type asyncByteSliceInvocation struct {
	Func       func() []byte
	ReturnChan chan *[]byte
}

/******************************************************************************
 New Functions
******************************************************************************/

func newAsyncVoidInvocation(f func()) *asyncVoidInvocation {
	return &asyncVoidInvocation{
		Func:     f,
		DoneChan: make(chan bool, 1),
	}
}

func newAsyncBoolInvocation(f func() bool) *asyncBoolInvocation {
	return &asyncBoolInvocation{
		Func:       f,
		ReturnChan: make(chan bool, 1),
	}
}

func newAsyncByteSliceInvocation(f func() []byte) *asyncByteSliceInvocation {
	return &asyncByteSliceInvocation{
		Func:       f,
		ReturnChan: make(chan *[]byte, 1),
	}
}
