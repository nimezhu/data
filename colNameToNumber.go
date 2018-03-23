package data

import (
	"strings"
)

func colNameToNumber(name string) int {
	if len(name) == 0 {
		return -1 //wrong
	}
	n := strings.ToUpper(name)
	sum := 0
	for i := 0; i < len(n); i++ {
		sum *= 26
		sum += int(n[i]) - int('A') + 1
	}
	return sum
}
