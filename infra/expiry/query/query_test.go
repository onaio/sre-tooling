package query

import (
	"testing"
	"time"

	"github.com/onaio/sre-tooling/libs/infra"
)

// Test whether maxAge returns the right values if the provided maximum age is a blank string
func TestHasMaxAgeReachedEmptyMaxAge(t *testing.T) {
	resource := infra.Resource{
		Provider:   "testProvider",
		ID:         "blah-id",
		Location:   "eu-central-1a",
		LaunchTime: time.Now(),
		Tags:       map[string]string{},
		Properties: map[string]string{}}
	maxAge := ""
	maxAgeReached, _, maxAgeErr := hasMaxAgeReached(&resource, &maxAge)

	if maxAgeReached {
		t.Errorf("Resource should not be expired if provided maximum age is '%s'", maxAge)
	}

	if maxAgeErr != nil {
		t.Errorf("Error should not be returned if provided maximum age is '%s'", maxAge)
	}
}

// Test whether maxAge returns the right values if the provided maximum age is an unparseable string
func TestHasMaxAgeReachedIncorrectMaxAge(t *testing.T) {
	resource := infra.Resource{
		Provider:   "testProvider",
		ID:         "blah-id",
		Location:   "eu-central-1a",
		LaunchTime: time.Now(),
		Tags:       map[string]string{},
		Properties: map[string]string{}}
	maxAge := "dfds"
	maxAgeReached, _, maxAgeErr := hasMaxAgeReached(&resource, &maxAge)

	if maxAgeReached {
		t.Errorf("Resource should not be expired if provided maximum age is '%s'", maxAge)
	}

	if maxAgeErr == nil {
		t.Errorf("Error should be returned if provided maximum age is '%s'", maxAge)
	}
}

// Test whether maxAge returns the right values if the provided maximum age hasn't yet been reached
func TestHasMaxAgeReachedImmatureMaxAge(t *testing.T) {
	resource := infra.Resource{
		Provider:   "testProvider",
		ID:         "blah-id",
		Location:   "eu-central-1a",
		LaunchTime: time.Now(),
		Tags:       map[string]string{},
		Properties: map[string]string{}}
	maxAge := "1h" // resource is expected to expire after 1 hour
	maxAgeReached, _, maxAgeErr := hasMaxAgeReached(&resource, &maxAge)

	if maxAgeReached {
		t.Errorf("Resource should not be expired if provided maximum age is '%s'", maxAge)
	}

	if maxAgeErr != nil {
		t.Errorf("Error should not be returned if provided maximum age is '%s'", maxAge)
	}
}

// Test whether maxAge returns the right values if the provided maximum age has been reached
func TestHasMaxAgeReachedMatureMaxAge(t *testing.T) {
	launchTime := time.Now().Add(-time.Hour) // 1 hour in the past
	resource := infra.Resource{
		Provider:   "testProvider",
		ID:         "blah-id",
		Location:   "eu-central-1a",
		LaunchTime: launchTime,
		Tags:       map[string]string{},
		Properties: map[string]string{}}
	maxAge := "40m" // resource is expected to expire after 30 min from now
	maxAgeReached, expiryTime, maxAgeErr := hasMaxAgeReached(&resource, &maxAge)

	if !maxAgeReached {
		t.Errorf("Resource should be expired if provided maximum age is '%s'", maxAge)
	}

	if maxAgeErr != nil {
		t.Errorf("Error should not be returned if provided maximum age is '%s'", maxAge)
	}

	// expecting expiry time to be around 20min in the past
	timeDiff := time.Now().Unix() - expiryTime.Unix()
	if timeDiff < (19*60) || timeDiff > (21*60) { // Give an error mergin of 2min
		t.Errorf("Expecting expiry time to be around 20min in the past if maximum age is '%s' and launch time is '%s'. Time difference is '%d' seconds", maxAge, launchTime.Format(time.RFC1123), timeDiff)
	}
}

// Test whether hasExpiryTimeMatured returns the right values if the provided tag is not set
func TestHasExpiryTimeMaturedTagNotSet(t *testing.T) {
	tagName := "someRandomTag"
	expiryTagNAValue := defaultExpiryTagNAValue
	timeFormat := defaultTimeFormat

	resource := infra.Resource{
		Provider:   "testProvider",
		ID:         "blah-id",
		Location:   "eu-central-1a",
		LaunchTime: time.Now(),
		Tags:       map[string]string{},
		Properties: map[string]string{}}

	expiryMatured, _, expiryMaturedErr := hasExpiryTimeMatured(&resource, &tagName, &expiryTagNAValue, &timeFormat)

	if expiryMatured {
		t.Errorf("Resource should not be expired if provided expiry tag has not been set")
	}

	if expiryMaturedErr != nil {
		t.Errorf("Error should not be returned if provided expiry tag has not been set")
	}
}

// Test whether hasExpiryTimeMatured returns the right values if the provided tag's value is empty
func TestHasExpiryTimeMaturedTagEmpty(t *testing.T) {
	tagName := "someRandomTag"
	expiryTagNAValue := defaultExpiryTagNAValue
	timeFormat := defaultTimeFormat

	resource := infra.Resource{
		Provider:   "testProvider",
		ID:         "blah-id",
		Location:   "eu-central-1a",
		LaunchTime: time.Now(),
		Tags: map[string]string{
			tagName: "",
		},
		Properties: map[string]string{}}

	expiryMatured, _, expiryMaturedErr := hasExpiryTimeMatured(&resource, &tagName, &expiryTagNAValue, &timeFormat)

	if expiryMatured {
		t.Errorf("Resource should not be expired if provided expiry tag is empty")
	}

	if expiryMaturedErr == nil {
		t.Errorf("Error should be returned if provided expiry tag has is empty")
	}
}

// Test whether hasExpiryTimeMatured returns the right values if the provided tag's value cannot be parsed
func TestHasExpiryTimeMaturedIncorrectTagValue(t *testing.T) {
	tagName := "someRandomTag"
	tagValue := "sdfsdadds"
	expiryTagNAValue := defaultExpiryTagNAValue
	timeFormat := defaultTimeFormat

	resource := infra.Resource{
		Provider:   "testProvider",
		ID:         "blah-id",
		Location:   "eu-central-1a",
		LaunchTime: time.Now(),
		Tags: map[string]string{
			tagName: tagValue,
		},
		Properties: map[string]string{}}

	expiryMatured, _, expiryMaturedErr := hasExpiryTimeMatured(&resource, &tagName, &expiryTagNAValue, &timeFormat)

	if expiryMatured {
		t.Errorf("Resource should not be expired if provided expiry tag is '%s'", tagValue)
	}

	if expiryMaturedErr == nil {
		t.Errorf("Error should be returned if provided expiry tag has is '%s'", tagValue)
	}
}

// Test whether hasExpiryTimeMatured returns the right values if the incorrect time format is provided
func TestHasExpiryTimeMaturedIncorrectTimeFormat(t *testing.T) {
	tagName := "someRandomTag"
	tagValue := "2020-01-01"
	expiryTagNAValue := defaultExpiryTagNAValue
	timeFormat := "fdsafdsfewf"

	resource := infra.Resource{
		Provider:   "testProvider",
		ID:         "blah-id",
		Location:   "eu-central-1a",
		LaunchTime: time.Now(),
		Tags: map[string]string{
			tagName: tagValue,
		},
		Properties: map[string]string{}}

	expiryMatured, _, expiryMaturedErr := hasExpiryTimeMatured(&resource, &tagName, &expiryTagNAValue, &timeFormat)

	if expiryMatured {
		t.Errorf("Resource should not be expired if expiry time format is '%s'", timeFormat)
	}

	if expiryMaturedErr == nil {
		t.Errorf("Error should be returned if provided expiry time format is '%s'", timeFormat)
	}
}

// Test whether hasExpiryTimeMatured returns the right values if the resource tag indicates an expired resource
func TestHasExpiryTimeMaturedExpiredResource(t *testing.T) {
	expiryTagNAValue := defaultExpiryTagNAValue
	timeFormat := time.RFC1123
	tagName := "someRandomTag"
	tagValue := time.Now().Add(-(time.Hour * 24)).Format(timeFormat)

	resource := infra.Resource{
		Provider:   "testProvider",
		ID:         "blah-id",
		Location:   "eu-central-1a",
		LaunchTime: time.Now(),
		Tags: map[string]string{
			tagName: tagValue,
		},
		Properties: map[string]string{}}

	expiryMatured, expiryTime, expiryMaturedErr := hasExpiryTimeMatured(&resource, &tagName, &expiryTagNAValue, &timeFormat)

	if !expiryMatured {
		t.Errorf("Resource should be expired if expiry time is '%s'", tagValue)
	}

	if expiryMaturedErr != nil {
		t.Errorf("Error should not be returned")
	}

	// expecting expiry time to be around 20min in the past
	timeDiff := time.Now().Unix() - expiryTime.Unix()
	if timeDiff < (23*3600) || timeDiff > (25*3600) { // Give an error mergin of 2 hours
		t.Errorf("Expecting expiry time to be around 24hrs in the past if expiry time is '%s'", tagValue)
	}
}

// Test whether hasExpiryTimeMatured returns the right values if the resource tag indicates a resource that hasn't expired
func TestHasExpiryTimeMaturedNotExpiredResource(t *testing.T) {
	expiryTagNAValue := defaultExpiryTagNAValue
	timeFormat := time.RFC1123
	tagName := "someRandomTag"
	tagValue := time.Now().Add((time.Hour)).Format(timeFormat)

	resource := infra.Resource{
		Provider:   "testProvider",
		ID:         "blah-id",
		Location:   "eu-central-1a",
		LaunchTime: time.Now(),
		Tags: map[string]string{
			tagName: tagValue,
		},
		Properties: map[string]string{}}

	expiryMatured, _, expiryMaturedErr := hasExpiryTimeMatured(&resource, &tagName, &expiryTagNAValue, &timeFormat)

	if expiryMatured {
		t.Errorf("Resource should not be expired if expiry time is '%s'", tagValue)
	}

	if expiryMaturedErr != nil {
		t.Errorf("Error should not be returned")
	}
}
