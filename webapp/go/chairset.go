package main

import (
	"strconv"
	"sync"

	"github.com/wangjia184/sortedset"
)

var chairlock sync.Mutex
var byPriceChair *sortedset.SortedSet

func AddToChairSet(chair Chair) {
	chairlock.Lock()
	defer chairlock.Unlock()
	key := strconv.Itoa(int(chair.ID))
	score := sortedset.SCORE(chair.Price*100000 + chair.ID)
	byPriceChair.AddOrUpdate(key, score, chair)
	if chair.Stock == 0 {
		byPriceChair.Remove(key)
	}
}

func ResetChairSet() {
	chairs := make([]Chair, 0)
	err := db_estate.Select(&chairs, "SELECT * FROM `chair`")
	if err != nil {
		panic(err)
	}
	byPriceChair = sortedset.New()
	for _, chair := range chairs {
		AddToChairSet(chair)
	}
}
