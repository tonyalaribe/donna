package lape

import (
        `math/rand`
        `time`
)

// Returns row number for the given bit index.
func Row(n int) int {
	return n / 8 // n >> 3
}

// Returns column number for the given bit index.
func Column(n int) int {
	return n % 8 // n & 7
}

// Returns row and column numbers for the given bit index.
func Coordinate(n int) (int, int) {
        return Row(n), Column(n)
}

// Returns n for the given the given row/column coordinate.
func Index(row, column int) int {
	return (row << 3) + column
}

// Integer version of math/abs.
func Abs(n int) int {
        if n < 0 {
                return -n
        }
        return n
}

func Random(limit int) int {
        rand.Seed(time.Now().Unix())
        return rand.Intn(limit)
}