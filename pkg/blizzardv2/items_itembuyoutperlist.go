package blizzardv2

import (
	"math"
	"sort"

	"gonum.org/v1/gonum/stat"
)

type ItemBuyoutPerList []float64

func (perList ItemBuyoutPerList) Average() float64 {
	total := float64(0)
	for _, buyout := range perList {
		total += buyout
	}

	return total / float64(len(perList))
}

func (perList ItemBuyoutPerList) Median() float64 {
	buyoutsSlice := sort.Float64Slice(perList)
	buyoutsSlice.Sort()
	hasEvenMembers := len(buyoutsSlice)%2 == 0
	if hasEvenMembers {
		middle := float64(len(buyoutsSlice)) / 2

		return (buyoutsSlice[int(math.Floor(middle))] + buyoutsSlice[int(math.Ceil(middle))]) / 2
	}

	return buyoutsSlice[(len(buyoutsSlice)-1)/2]
}

func (perList ItemBuyoutPerList) StdDev() float64 {
	return stat.StdDev(perList, nil)
}
