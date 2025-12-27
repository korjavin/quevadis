package main

import (
	"math/rand"
	"strconv"
	"time"
)

var (
	adjectives = []string{
		"Brave", "Clever", "Swift", "Bright", "Calm",
		"Keen", "Bold", "Wise", "Quick", "Fair",
		"Kind", "True", "Fine", "Grand", "Happy",
		"Merry", "Noble", "Proud", "Safe", "Warm",
	}
	animals = []string{
		"Badger", "Bear", "Beaver", "Coyote", "Eagle",
		"Fox", "Hawk", "Lion", "Owl", "Wolf",
		"Boar", "Deer", "Elk", "Swan", "Seal",
		"Whale", "Otter", "Lynx", "Viper", "Tiger",
	}
)

func GenerateRandomName() string {
	adj := adjectives[rand.Intn(len(adjectives))]
	animal := animals[rand.Intn(len(animals))]
	number := rand.Intn(1000)
	return adj + animal + strconv.Itoa(number)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
