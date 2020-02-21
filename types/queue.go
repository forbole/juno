package types

// Queue is a simple type alias for a (buffered) channel of block heights.
type Queue chan int64

func NewQueue(size int) Queue {
	return make(chan int64, size)
}
