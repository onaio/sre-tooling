package numbers

import (
	"fmt"
	"math"
	"testing"
)

// Test whether 0 is always returned if maximum random number is set to 0
func TestGetRandomIntZeroIsMax(t *testing.T) {
	for i := 0; i < 1000; i++ {
		randomInt := GetRandomInt(0)
		if randomInt != 0 {
			t.Errorf("random integer = %d; want 0", randomInt)
		}
	}
}

// Test whether the getRandomInt function actually returns an acceptable
// ratio of random numbers
func TestGetRandomIntMaxMoreThanZero(t *testing.T) {
	generatedNumbers := make(map[int]int)
	numberOfRandomInts := 10
	maxRandomValue := 20 // random numbers will be between 0 and maxRandomValue
	for i := 0; i < numberOfRandomInts; i++ {
		randomInt := GetRandomInt(maxRandomValue)
		noOccurrences, intInMap := generatedNumbers[randomInt]
		if !intInMap {
			noOccurrences = 0
		}
		noOccurrences++
		generatedNumbers[randomInt] = noOccurrences
	}

	// Use the birthday problem's solution to determine if the number
	// of collisions (situations where random int gotten more than once)
	// are acceptable
	// https://en.wikipedia.org/wiki/Birthday_problem
	//
	// probability of collision = (maxRandomValue P numberOfRandomInts) / maxRandomValue ^ numberOfRandomInts

	probability := Permutation(maxRandomValue, numberOfRandomInts) / math.Pow(float64(maxRandomValue), float64(numberOfRandomInts))
	fmt.Printf("Probability is %f\n", probability)

	// Multiply by 4 because we expect the random number generator to be pseudorandom
	// hence more likely to lead to collisions.
	// Just for illustration, in cases where:
	//   maxRandomValue = 20
	//   numberOfRandomInts = 10
	//   probability of collision = 0.065473
	//   statistical maximum acceptable collisions = 1 (highly unlikely we can achieve this with pseudorandomness)
	//   acceptable number of collisions = 4
	maxAcceptableCollisions := int(math.Round(probability*float64(numberOfRandomInts))) * 4
	numberOfCollisions := 0
	maxOccurrences := 0 // the highest number of occurrences of any of the random numbers
	for _, numberOccurrences := range generatedNumbers {
		if numberOccurrences > 1 {
			numberOfCollisions++
		}

		if maxOccurrences < numberOccurrences {
			maxOccurrences = numberOccurrences
		}

		if numberOfCollisions > maxAcceptableCollisions {
			t.Errorf(
				"Number of collisions %d exceeded maximum acceptable number %d",
				numberOfCollisions,
				maxAcceptableCollisions)
			break
		}
		if maxOccurrences == numberOfRandomInts {
			t.Errorf("Radomness not achieved at all. One digit regenerate %d times\n", maxOccurrences)
			break
		}
	}
	if maxAcceptableCollisions >= numberOfRandomInts {
		t.Errorf(
			"Not expecting maximum acceptable number of collisions (%d) >= number of generated ints (%d)",
			maxAcceptableCollisions,
			numberOfRandomInts)
	}
}
