package calculate

import (
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/onaio/sre-tooling/libs/cloud"
)

// Test new index when two resources with duplicate indexes provided
func TestGetNewResourceIndexDuplicateIndex(t *testing.T) {
	// Test new index when two resources with duplicate indexes provided
	resource1Id := "resource1"
	indexTag := "indexTag"
	index := "0"
	resource1 := cloud.Resource{
		Provider:   "testProvider",
		ID:         resource1Id,
		Location:   "eu-central-1a",
		LaunchTime: time.Now(),
		Tags:       map[string]string{indexTag: index},
		Properties: map[string]string{}}
	resource2 := cloud.Resource{
		Provider:   "testProvider",
		ID:         "resource2",
		Location:   "eu-central-1a",
		LaunchTime: time.Now(),
		Tags:       map[string]string{indexTag: index},
		Properties: map[string]string{}}
	calculatedIndex, err := GetNewResourceIndex(
		&resource1Id,
		&indexTag,
		[]*cloud.Resource{&resource1, &resource2})
	if calculatedIndex != 1 {
		t.Errorf("Index for resource1 = %d; want 1", calculatedIndex)
	}
	if err != nil {
		t.Errorf("Error for resource1 = '%s'; want nil", err.Error())
	}
}

// Test new index when two resources with duplicate indexes provided
// and index greater than 0
func TestGetNewResourceIndexDuplicateIndexGreaterZero(t *testing.T) {
	resource1Id := "resource1"
	indexTag := "indexTag"
	index := "2"
	resource1 := cloud.Resource{
		Provider:   "testProvider",
		ID:         resource1Id,
		Location:   "eu-central-1a",
		LaunchTime: time.Now(),
		Tags:       map[string]string{indexTag: index},
		Properties: map[string]string{}}
	resource2 := cloud.Resource{
		Provider:   "testProvider",
		ID:         "resource2",
		Location:   "eu-central-1a",
		LaunchTime: time.Now(),
		Tags:       map[string]string{indexTag: index},
		Properties: map[string]string{}}
	calculatedIndex, err := GetNewResourceIndex(
		&resource1Id,
		&indexTag,
		[]*cloud.Resource{&resource1, &resource2})
	if calculatedIndex != 0 {
		t.Errorf("Index for resource1 = %d; want 1", calculatedIndex)
	}
	if err != nil {
		t.Errorf("Error for resource1 = '%s'; want nil", err.Error())
	}
}

// Test if error thrown when provided resource ID is not found
func TestGetNewResourceIndexResourceNotFound(t *testing.T) {
	resource1Id := "resource1"
	indexTag := "indexTag"
	resource1 := cloud.Resource{
		Provider:   "testProvider",
		ID:         resource1Id,
		Location:   "eu-central-1a",
		LaunchTime: time.Now(),
		Tags:       map[string]string{indexTag: "0"},
		Properties: map[string]string{}}
	_, err := GetNewResourceIndex(
		&resource1Id,
		&indexTag,
		[]*cloud.Resource{&resource1})
	if err == nil {
		t.Errorf("Error for resource1 = nil; want != nil")
	}
}

// Test resources that don't need to be assigned new indexes (because their
// current index is unique)
func TestGetNewResourceIndexIndexOk(t *testing.T) {
	resource1Id := "resource1"
	indexTag := "indexTag"
	resource1Index := 0
	resource1 := cloud.Resource{
		Provider:   "testProvider",
		ID:         resource1Id,
		Location:   "eu-central-1a",
		LaunchTime: time.Now(),
		Tags:       map[string]string{indexTag: strconv.Itoa(resource1Index)},
		Properties: map[string]string{}}
	resource2Index := 1
	resource2Id := "resource2"
	resource2 := cloud.Resource{
		Provider:   "testProvider",
		ID:         resource2Id,
		Location:   "eu-central-1a",
		LaunchTime: time.Now(),
		Tags:       map[string]string{indexTag: strconv.Itoa(resource2Index)},
		Properties: map[string]string{}}
	calculatedIndex, err := GetNewResourceIndex(
		&resource1Id,
		&indexTag,
		[]*cloud.Resource{&resource1, &resource2})
	if calculatedIndex != resource1Index {
		t.Errorf("Index for resource1 = %d; want %d", calculatedIndex, resource1Index)
	}
	if err == nil {
		t.Errorf("Error for resource1 = nil; want != nil")
	}
	calculatedIndex, err = GetNewResourceIndex(
		&resource2Id,
		&indexTag,
		[]*cloud.Resource{&resource1, &resource2})
	if calculatedIndex != resource2Index {
		t.Errorf("Index for resource2 = %d; want %d", calculatedIndex, resource2Index)
	}
	if err == nil {
		t.Errorf("Error for resource2 = nil; want != nil")
	}
}

// Test if 0 gotten if index tag is an empty string
func TestGetResourceIndexEmptyString(t *testing.T) {
	indexTag := "indexTag"
	resource1 := cloud.Resource{
		Tags: map[string]string{indexTag: ""}}
	index, err := getResourceIndex(&resource1, &indexTag)
	if err != nil {
		t.Errorf("Error for getResourceIndex() = '%s'; want nil", err.Error())
	}
	if index != 0 {
		t.Errorf("Index = '%d'; want '%d'", index, 0)
	}
}

// Test if 0 gotten if index tag not set
func TestGetResourceIndexTagNotSet(t *testing.T) {
	indexTag := "indexTag"
	resource1 := cloud.Resource{
		Tags: map[string]string{}}
	index, err := getResourceIndex(&resource1, &indexTag)
	if err != nil {
		t.Errorf("Error for getResourceIndex() = '%s'; want nil", err.Error())
	}
	if index != 0 {
		t.Errorf("Index = '%d'; want '%d'", index, 0)
	}
}

// Test if error gotten if index tag value is not a number
func TestGetResourceIndexNonNumberIndex(t *testing.T) {
	indexTag := "indexTag"
	resource1 := cloud.Resource{
		Tags: map[string]string{indexTag: "dfd"}}
	_, err := getResourceIndex(&resource1, &indexTag)
	if err == nil {
		t.Errorf("Error for getResourceIndex() = nil; want != nil")
	}
}

// Test if right number gotten if index tag has a number
func TestGetResourceIndexOk(t *testing.T) {
	indexTag := "indexTag"
	indexString := "231232"
	resource1 := cloud.Resource{
		Tags: map[string]string{indexTag: indexString}}
	index, err := getResourceIndex(&resource1, &indexTag)
	if err != nil {
		t.Errorf("Error for getResourceIndex() = %s; want nil", err.Error())
	}
	indexStrVal, _ := strconv.Atoi(indexString)
	if indexStrVal != index {
		t.Errorf("Index = %d; want %d", index, indexStrVal)
	}
}

// Test whether 0 is always returned if maximum random number is set to 0
func TestGetRandomIntZeroIsMax(t *testing.T) {
	for i := 0; i < 1000; i++ {
		randomInt := getRandomInt(0)
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
		randomInt := getRandomInt(maxRandomValue)
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

	probability := permutation(maxRandomValue, numberOfRandomInts) / math.Pow(float64(maxRandomValue), float64(numberOfRandomInts))
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

func factorial(n int) uint64 {
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

func permutation(n int, k int) float64 {
	return float64(factorial(n)) / float64(factorial(n-k))
}
