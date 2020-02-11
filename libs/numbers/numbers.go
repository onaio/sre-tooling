package numbers

import "math/rand"

// GetRandomInt returns a random integer between 1 and maxPossibleValue. If maxPossibleValue is
// less than or equal to 0, then 0 is returned
func GetRandomInt(maxPossibleValue int) int {
	if maxPossibleValue > 0 {
		return rand.Intn(maxPossibleValue)
	}

	return 0
}

// Factorial calculates the factorial of the provided integer
func Factorial(n int) uint64 {
	var factVal uint64 = 1
	if n < 0 {
		panic("Factorial of negative number doesn't exist.")
	} else {
		for i := 1; i <= n; i++ {
			factVal *= uint64(i)
		}

	}
	return factVal
}

// Permutation calculates the permutation as n P r
// where n is the number of choices for the items
// and r the number of ordered items
func Permutation(n int, r int) float64 {
	return float64(Factorial(n)) / float64(Factorial(n-r))
}
