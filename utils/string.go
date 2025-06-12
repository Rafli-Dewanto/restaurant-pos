package utils

import (
	"math"
	"math/rand"
)

func RandomNumber(digits int) int {
	return rand.Intn(int(math.Pow10(digits)))
}
