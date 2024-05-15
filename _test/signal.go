package _test

import (
	"github.com/tonybillings/gfx"
	"math/rand"
)

/******************************************************************************
 OddValueFilter
******************************************************************************/

type OddValueFilter struct {
	gfx.FilterBase
}

func (f *OddValueFilter) Apply(index int, input []float64) (output float64) {
	if int(input[index])%2 == 0 {
		return input[index]
	} else {
		return input[index] + 1
	}
}

func NewOddValueFilter() *OddValueFilter {
	f := &OddValueFilter{}
	f.SetEnabled(true)
	return f
}

/******************************************************************************
 PlusNTransformer
******************************************************************************/

type PlusNTransformer struct {
	gfx.TransformerBase
}

func (t *PlusNTransformer) Transform(dst, src []float64) []float64 {
	n := rand.Float64() * 10
	nArr := make([]float64, len(src))
	for i, v := range src {
		dst[i] = v + n
		nArr[i] = n
		n = rand.Float64() * 10
	}
	return nArr
}

func NewPlusNTransformer() *PlusNTransformer {
	t := &PlusNTransformer{}
	t.SetEnabled(true)
	return t
}
