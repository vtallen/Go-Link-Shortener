// File: pkg/codegen/codegen.go
// Includes utilities for generating random IDs and converting between base-10 and base-N representations
// This gets used to create shortcodes for a database

package codegen

import (
	"math"
	"math/rand"
)

/*
* Function: GenRandID
*
* Parameters: universe string - The set of characters to use when generating the random ID
*             maxchars int - The maximum number of characters in the generated ID
*
* Returns: int - The generated random ID
*
* Description: Generates a random ID using the given universe of characters and maximum number of characters
 */
func GenRandID(universe string, maxchars int) int {
	max_result := ""
	for idx := 0; idx < maxchars; idx++ {
		max_result += string(universe[len(universe)-1])
	}
	max_index := UniverseToBaseTen(max_result, universe)

	result := rand.Intn(max_index)

	return result
}

/*
* Function: BaseTenToUniverse
*
* Parameters: baseten int - The base-10 number to convert
*             universe string - The set of characters to use when converting to base-N
*
* Returns: string - The base-N representation of the base-10 number where n is the length of the universe string
*
* Description: Converts a base-10 number to a base-N number using the given universe of characters
 */
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

/*
* Function: UniverseToBaseTen
*
* Parameters: input string - The base-N number to convert
*             universe string - The set of characters to use when converting to base-10
*
* Returns: int - The base-10 representation of the base-N number where n is the length of the universe string
*
* Description: Converts a base-N number to a base-10 number using the given universe of characters
 */
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
