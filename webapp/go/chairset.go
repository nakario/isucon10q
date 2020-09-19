package main

import (
	"log"
	"strconv"
	"sync"

	"github.com/wangjia184/sortedset"
)

// key: height(4+1), width(4+1), depth(4+1)
//      price(6+1), color(12+1), kind(4+1)
// * 8倍のinsert/メモリ
var chairSet [][][][][][](*sortedset.SortedSet)
var chairlock sync.Mutex
var byPriceChair *sortedset.SortedSet

func PriceToIndex(price int64) int64 {
	if price < 3000 {
		return 0
	}
	if price < 6000 {
		return 1
	}
	if price < 9000 {
		return 2
	}
	if price < 12000 {
		return 3
	}
	if price < 15000 {
		return 4
	}
	return 5
}
func ColorToIndex(color string) int64 {
	if color == "黒" {
		return 0
	}
	if color == "白" {
		return 1
	}
	if color == "赤" {
		return 2
	}
	if color == "青" {
		return 3
	}
	if color == "緑" {
		return 4
	}
	if color == "黄" {
		return 5
	}
	if color == "紫" {
		return 6
	}
	if color == "ピンク" {
		return 7
	}
	if color == "オレンジ" {
		return 8
	}
	if color == "水色" {
		return 9
	}
	if color == "ネイビー" {
		return 10
	}
	if color == "ベージュ" {
		return 11
	}
	return -1
}
func KindToIndex(kind string) int64 {
	if kind == "ゲーミングチェア" {
		return 0
	}
	if kind == "座椅子" {
		return 1
	}
	if kind == "エルゴノミクス" {
		return 2
	}
	if kind == "ハンモック" {
		return 3
	}
	return -1
}

// key: height(4+1), width(4+1), depth(4+1)
//      price(6+1), color(12+1), kind(4+1)

func QueryChairSet(
	height int64, width int64, depth int64, price int64,
	color string, colorIsAll bool, kind string, kindIsAll bool,
	limit int64, offset int64) []Chair {
	h := int64(4)
	w := int64(4)
	d := int64(4)
	p := int64(6)
	c := int64(12)
	k := int64(4)
	if height >= 0 {
		h = SizeToIndex(height)
	}
	if width >= 0 {
		w = SizeToIndex(width)
	}
	if depth >= 0 {
		d = SizeToIndex(depth)
	}
	if price >= 0 {
		p = PriceToIndex(price)
	}
	if !colorIsAll {
		c = ColorToIndex(color)
		if c < 0 {
			return nil
		}
	}
	if !kindIsAll {
		k = KindToIndex(kind)
		if k < 0 {
			return nil
		}
	}
	result := make([]Chair, 0)
	ids := chairSet[h][w][d][p][c][k].GetByRankRange(int(-1-offset), int(-1-offset-limit), false)
	for _, id := range ids {
		if id == nil {
			continue
		}
		chair := GetChair(id.Value.(int64))
		if chair == nil {
			continue
		}
		result = append(result, chair.Value.(Chair))
	}
	return result
}
func GetChair(id int64) *sortedset.SortedSetNode {
	return byPriceChair.GetByKey(strconv.Itoa(int(id)))
}

func AddToChairSet(chair Chair) bool {
	score := sortedset.SCORE(chair.Popularity*100000 - chair.ID)
	h := SizeToIndex(chair.Height)
	w := SizeToIndex(chair.Width)
	d := SizeToIndex(chair.Depth)
	p := PriceToIndex(chair.Price)
	c := ColorToIndex(chair.Color)
	k := KindToIndex(chair.Kind)
	key := strconv.Itoa(int(chair.ID))
	if c < 0 || k < 0 {
		log.Printf("Invalid chair!!")
	}
	chairlock.Lock()
	defer chairlock.Unlock()
	for _, hi := range []int64{h, 4} {
		for _, wi := range []int64{w, 4} {
			for _, di := range []int64{d, 4} {
				for _, pi := range []int64{p, 6} {
					for _, ci := range []int64{c, 12} {
						for _, ki := range []int64{k, 4} {
							if chair.Stock == 0 {
								chairSet[hi][wi][di][pi][ci][ki].Remove(key)
							} else {
								chairSet[hi][wi][di][pi][ci][ki].AddOrUpdate(key, score, chair.ID)
							}
						}
					}
				}
			}
		}
	}

	priceScore := sortedset.SCORE(chair.Price*100000 + chair.ID)
	if chair.Stock == 0 {
		byPriceChair.Remove(key)
	} else {
		byPriceChair.AddOrUpdate(key, priceScore, chair)
	}
	return true
}

func ResetChairSet() {
	chairSet = make([][][][][][](*sortedset.SortedSet), 5)
	for h := 0; h < 5; h++ {
		chairSet[h] = make([][][][][](*sortedset.SortedSet), 5)
		for w := 0; w < 5; w++ {
			chairSet[h][w] = make([][][][](*sortedset.SortedSet), 5)
			for d := 0; d < 5; d++ {
				chairSet[h][w][d] = make([][][](*sortedset.SortedSet), 7)
				for p := 0; p < 7; p++ {
					chairSet[h][w][d][p] = make([][](*sortedset.SortedSet), 13)
					for c := 0; c < 13; c++ {
						chairSet[h][w][d][p][c] = make([](*sortedset.SortedSet), 5)
						for k := 0; k < 5; k++ {
							chairSet[h][w][d][p][c][k] = sortedset.New()
						}
					}
				}
			}
		}
	}
	chairs := make([]Chair, 0)
	err := db_chair.Select(&chairs, "SELECT * FROM `chair`")
	if err != nil {
		panic(err)
	}
	byPriceChair = sortedset.New()
	for _, chair := range chairs {
		AddToChairSet(chair)
	}
}
