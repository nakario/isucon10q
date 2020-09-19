package main

import (
	"strconv"
	"sync"

	"github.com/wangjia184/sortedset"
)

// key: rent(4+1) width(4+1) height(4+1)
var estateSet [][][](*sortedset.SortedSet)
var estatelock sync.Mutex
var byRentEState *sortedset.SortedSet

func SizeToIndex(size int64) int64 {
	if size < 80 {
		return 0
	}
	if size < 110 {
		return 1
	}
	if size < 150 {
		return 2
	}
	return 3
}

func RentToIndex(rent int64) int64 {
	if rent < 50000 {
		return 0
	}
	if rent < 100000 {
		return 1
	}
	if rent < 150000 {
		return 2
	}
	return 3
}

func AddToEstateSet(estate Estate) {
	score := sortedset.SCORE(estate.Popularity*100000 - estate.ID)
	r := RentToIndex(estate.Rent)
	w := SizeToIndex(estate.DoorWidth)
	h := SizeToIndex(estate.DoorHeight)
	key := strconv.Itoa(int(estate.ID))
	estatelock.Lock()
	defer estatelock.Unlock()
	for _, ri := range []int64{r, 4} {
		for _, wi := range []int64{w, 4} {
			for _, hi := range []int64{h, 4} {
				estateSet[ri][wi][hi].AddOrUpdate(key, score, estate.ID)
			}
		}
	}
	rentScore := sortedset.SCORE(estate.Rent*100000 + estate.ID)
	byRentEState.AddOrUpdate(key, rentScore, estate)
}
func GetEstate(id int64) *sortedset.SortedSetNode {
	return byRentEState.GetByKey(strconv.Itoa(int(id)))
}

func GetEstateSetCount(rent int64, width int64, height int64) int64 {
	r := int64(4)
	w := int64(4)
	h := int64(4)
	if rent >= 0 {
		r = RentToIndex(rent)
	}
	if width >= 0 {
		w = SizeToIndex(width)
	}
	if height >= 0 {
		h = SizeToIndex(height)
	}
	return int64(estateSet[r][w][h].GetCount())
}

func QueryEstateSet(rent int64, width int64, height int64, limit int64, offset int64) []Estate {
	r := int64(4)
	w := int64(4)
	h := int64(4)
	if rent >= 0 {
		r = RentToIndex(rent)
	}
	if width >= 0 {
		w = SizeToIndex(width)
	}
	if height >= 0 {
		h = SizeToIndex(height)
	}
	result := make([]Estate, 0)
	ids := estateSet[r][w][h].GetByRankRange(int(-1-offset), int(-1-offset-limit), false)
	for _, id := range ids {
		if ids == nil {
			continue
		}
		estate := GetEstate(id.Value.(int64))
		if estate == nil {
			continue
		}
		result = append(result, estate.Value.(Estate))
	}
	return result
}

func ResetEstateSet() {
	estateSet = make([][][](*sortedset.SortedSet), 5)
	for r := 0; r < 5; r++ {
		estateSet[r] = make([][](*sortedset.SortedSet), 5)
		for w := 0; w < 5; w++ {
			estateSet[r][w] = make([](*sortedset.SortedSet), 5)
			for h := 0; h < 5; h++ {
				estateSet[r][w][h] = sortedset.New()
			}
		}
	}
	estates := make([]Estate, 0)
	err := db_estate.Select(&estates, "SELECT id, name, description, thumbnail, address, latitude, longitude, rent, door_height, door_width, features, popularity FROM `estate`")
	if err != nil {
		panic(err)
	}
	byRentEState = sortedset.New()
	for _, estate := range estates {
		AddToEstateSet(estate)
	}
}
