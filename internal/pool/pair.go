package pool

type Pair struct {
	Name  string
	Value any
}

var pairSlicePool = SlicePool[Pair]{
	pool: New[[]Pair](allocPairSlice, freePairSlice),
}

func allocPairSlice() []Pair {
	return make([]Pair, 0, 64) // Initial capacity of 64
}

func freePairSlice(slice []Pair) []Pair {
	clear(slice)
	slice = slice[:0] // Reset length to 0
	return slice
}

func PairSlice() SlicePool[Pair] {
	return pairSlicePool
}
