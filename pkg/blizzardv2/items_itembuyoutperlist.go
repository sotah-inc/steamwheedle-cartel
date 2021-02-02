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

func (perList ItemBuyoutPerList) Sort() ItemBuyoutPerList {
	buyoutsSlice := sort.Float64Slice(perList)
	buyoutsSlice.Sort()

	out := make([]float64, buyoutsSlice.Len())
	for i, per := range buyoutsSlice {
		out[i] = per
	}

	return out
}

func (perList ItemBuyoutPerList) CollectMinimum() ItemBuyoutPerList {
	count := float64(len(perList))

	if count <= 3 {
		return perList
	}

	out := ItemBuyoutPerList{}
	for i, per := range perList {
		completion := float64(i+1) / count
		if i == 0 || completion <= 0.15 {
			out = append(out, per)

			continue
		}

		if completion > 0.3 {
			break
		}

		prev := perList[i-1]
		prevRatio := per / prev
		if prevRatio >= 1.2 {
			break
		}

		out = append(out, per)
	}

	return out
}

func (perList ItemBuyoutPerList) RemoveOutliers() ItemBuyoutPerList {
	mean := stat.Mean(perList, nil)
	stdDev := stat.StdDev(perList, nil)
	stdDiffLimit := stdDev * 1.5

	out := ItemBuyoutPerList{}
	for _, per := range perList {
		diffFromAverage := math.Abs(mean - per)
		if diffFromAverage > stdDiffLimit {
			continue
		}

		out = append(out, per)
	}

	if len(out) == 0 {
		return perList
	}

	return out
}

func (perList ItemBuyoutPerList) MarketPrice() float64 {
	return stat.Mean(perList.Sort().CollectMinimum().RemoveOutliers(), nil)
}
