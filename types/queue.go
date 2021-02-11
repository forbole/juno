package types

// HeightQueue is a simple type alias for a (buffered) channel of block heights.
type HeightQueue chan int64

func NewQueue(size int) HeightQueue {
	return make(chan int64, size)
}
