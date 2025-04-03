package main

import (
	"math"
)

func mathCeilInInt(n, d int) int {
	if n%d == 0 {
		return n / d
	}
	return (n / d) + 1
}

func getLayout(groups ...int) (rows int, columns []int) {
	maxRows := 0
	for _, count := range groups {
		if count > maxRows {
			maxRows = count
		}
	}

	bestDiff := math.MaxFloat64
	bestRows := 0
	bestColumns := make([]int, len(groups))

	for r := 1; r <= maxRows; r++ {
		currentColumns := make([]int, len(groups))
		totalCols := 0

		for i, count := range groups {
			col := mathCeilInInt(count, r)
			currentColumns[i] = col
			totalCols += col
		}

		diff := math.Abs(float64(totalCols*9 - r*16))

		if diff < bestDiff {
			bestDiff = diff
			bestRows = r
			copy(bestColumns, currentColumns)
		}
	}

	return bestRows, bestColumns
}
