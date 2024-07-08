// Includes utilities for generating random IDs and converting between base-10 and base-N representations
// This gets used to create shortcodes for a database

package codegen

import (
	"math"
	"math/rand"
)

func GenRandID(universe string, maxchars int) int {
	max_result := ""
	for idx := 0; idx < maxchars; idx++ {
		max_result += string(universe[len(universe)-1])
	}
	max_index := UniverseToBaseTen(max_result, universe)

	result := rand.Intn(max_index)

	return result
}

func BaseTenToUniverse(baseten int, universe string) string {
	var base int = len(universe)
	var digits []int
	for baseten > 0 {
		digit := baseten % base
		digits = append(digits, digit)
		baseten = int(math.Floor(float64(baseten) / float64(base)))
	}

	result := ""
	for idx := len(digits) - 1; idx >= 0; idx-- {
		result += string(universe[digits[idx]])
	}

	return result
}

func UniverseToBaseTen(input string, universe string) int {
	base := len(universe)
	result := 0

	// Iterate over the input string in reverse order
	for idx := len(input) - 1; idx >= 0; idx-- {
		// Find the index of the character in the universe string
		char := input[idx]
		digit := 0
		for j := 0; j < len(universe); j++ {
			if universe[j] == char {
				digit = j
				break
			}
		}
		// Multiply the digit by the appropriate power of the base and add to result
		result += digit * int(math.Pow(float64(base), float64(len(input)-1-idx)))
	}

	return result
}
