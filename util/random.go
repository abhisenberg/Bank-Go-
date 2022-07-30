package util

import (
	"math/rand"
	"strings"
	"time"
)

var alphabets = "abcdefghijklmnopqrstuvxyz"

func init() {
	rand.Seed(time.Now().UnixNano())
}

//Generates a random number between min and max
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

//Generates a random string of length n
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabets)

	for i := 0; i < n; i++ {
		c := alphabets[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

//Genreate a random owner name
func RandomOwner() string {
	return RandomString(6)
}

//Generate a random amount of money
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

//Generate a random currency
func RandomCurrency() string {
	currencies := []string{"EUR", "USD", "CAD"}
	return currencies[rand.Intn(len(currencies))]
}
