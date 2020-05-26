package ingest

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

// Test whether the getLastIDStoragePath function works as expected
func TestGetLastIDStoragePath(t *testing.T) {
	for i := 0; i < 8; i++ {
		// For all decimal numbers between 0 and 7 get their binary
		// representation (should be 0 to 111) and use that to determine
		// if storageDir, flowBulletinURL, and sentryDSN should be set
		iBinary := fmt.Sprintf("%b", i)
		storageDir := ""
		if string(iBinary[0]) == "1" {
			storageDir = "/tmp"
		}
		flowBulletinURL := ""
		if len(iBinary) > 1 && string(iBinary[1]) == "1" {
			flowBulletinURL = "https://nifi.local/nifi-api/flow/bulletin-board"
		}
		sentryDSN := ""
		if len(iBinary) > 2 && string(iBinary[2]) == "1" {
			sentryDSN = "https://wr23423we23423423423@sentry.local/1"
		}

		path, pathErr := getLastIDStoragePath(&storageDir, &flowBulletinURL, &sentryDSN)
		if i == 7 { // 7 represented as binary is 111. Expected all the arguments to have been set
			if pathErr != nil {
				t.Errorf("Expecting error to be nil; got %v", pathErr)
			}

			// expected sha256 for flowBulletinURL + sentryDSN gotten using the
			// sha256sum command
			expectedPath := "/tmp/bbeb8a96b68bb71c83a0b3c0fd383a481c4dda811f6d9203beb77f574e27711b" + lastIDExtension
			if path != expectedPath {
				t.Errorf("Expected path to be '%s'; got '%s'", expectedPath, path)
			}
		} else if pathErr == nil {
			t.Errorf("Expecting error not to be nil; got %v", pathErr)
		}
	}
}

// Test whether the saveLastID function works as expected
func TestSaveLastID(t *testing.T) {
	storageDir1 := "."
	flowBulletinURL := "https://nifi.local/nifi-api/flow/bulletin-board"
	sentryDSN := "https://wr23423we23423423423@sentry.local/1"
	lastID := int64(234234234234234234)

	// Test 1: Last ID should be saved successfully
	saveErr1 := saveLastID(&storageDir1, &flowBulletinURL, &sentryDSN, lastID)
	path, _ := getLastIDStoragePath(&storageDir1, &flowBulletinURL, &sentryDSN)
	if saveErr1 != nil {
		t.Errorf("Expecting error to be nil; got %v", saveErr1)
	}
	fileContents, contentsErr := ioutil.ReadFile(path)
	if contentsErr != nil {
		t.Errorf("Expecting error to be nil; got %v", contentsErr)
	}
	if string(fileContents) != fmt.Sprintf("%d", lastID) {
		t.Errorf("Expecting file contest to be '%d'; got '%s'", lastID, fileContents)
	}

	// Test 2: saveLastID should return an error because storage dir doesn't exist
	storageDir2 := "/nonexisteddir_dsfsda"
	saveErr2 := saveLastID(&storageDir2, &flowBulletinURL, &sentryDSN, lastID)
	if saveErr2 == nil {
		t.Errorf("Expecting error not to be nil; got %v", saveErr2)
	}
}

// Benchmark how saveLastID performs
func BenchmarkSaveLastID(b *testing.B) {
	storageDir := "."
	flowBulletinURL := "https://nifi.local/nifi-api/flow/bulletin-board"
	sentryDSN := "https://wr23423we23423423423@sentry.local/1"
	lastID := int64(234234234234234234)

	for i := 0; i < b.N; i++ {
		saveLastID(&storageDir, &flowBulletinURL, &sentryDSN, lastID)
	}
}

// Test whether the getLastID function works as expected
func TestGetLastID(t *testing.T) {
	// Test 1: Return 0 if file doesn't exist
	storageDir := "."
	flowBulletinURL := "https://nifi.local/nifi-api/flow/bulletin-board"
	sentryDSN1 := "https://wr23423we23423423423@sentry.local/2"

	lastID1, lastIDErr1 := getLastID(&storageDir, &flowBulletinURL, &sentryDSN1)
	if lastID1 != 0 {
		t.Errorf("Expecting last ID to be 0; got %d", lastID1)
	}
	if lastIDErr1 != nil {
		t.Errorf("Expecting last ID error to be nil; got %v", lastIDErr1)
	}

	// Test 2: Return whatever saveLastID set
	sentryDSN2 := "https://wr23423we23423423423@sentry.local/3"
	lastID2 := int64(2342234234234222324)

	saveLastID(&storageDir, &flowBulletinURL, &sentryDSN2, lastID2)
	newLastID2, lastIDErr2 := getLastID(&storageDir, &flowBulletinURL, &sentryDSN2)
	if lastID2 != newLastID2 {
		t.Errorf("Expecting last ID to be %d; got %d", lastID2, newLastID2)
	}
	if lastIDErr2 != nil {
		t.Errorf("Expecting last ID error to be nil; got %v", lastIDErr2)
	}

	// Test 3: Error if contents of file is not a number
	sentryDSN3 := "https://wr23423we23423423423@sentry.local/4"
	lastID3 := "notanumber"
	path, _ := getLastIDStoragePath(&storageDir, &flowBulletinURL, &sentryDSN3)
	ioutil.WriteFile(path, []byte(lastID3), 0600)
	_, lastIDErr3 := getLastID(&storageDir, &flowBulletinURL, &sentryDSN3)
	if lastIDErr3 == nil {
		t.Errorf("Expecting error not to be nil; got %v", lastIDErr3)
	}
}

// Benchmark how getLastID performs
func BenchmarkGetLastID(b *testing.B) {
	storageDir := "."
	flowBulletinURL := "https://nifi.local/nifi-api/flow/bulletin-board"
	sentryDSN := "https://wr23423we23423423423@sentry.local/1"

	for i := 0; i < b.N; i++ {
		getLastID(&storageDir, &flowBulletinURL, &sentryDSN)
	}
}

// Test whether setLastID and getLastID will always write to and read from
// the right file
func TestUniqueLastIDSetGet(t *testing.T) {
	// Test 1: Unique flow bulletin URL
	for i := 10; i < 20; i++ {
		storageDir := "."
		flowBulletinURL := fmt.Sprintf("https://nifi.local/nifi-api/flow/bulletin-board/%d", i)
		sentryDSN := "https://wr23423we23423423423@sentry.local/5"
		saveLastID(&storageDir, &flowBulletinURL, &sentryDSN, int64(i))

		// Should be i
		lastID2, lastIDErr2 := getLastID(&storageDir, &flowBulletinURL, &sentryDSN)
		if lastID2 != int64(i) {
			t.Errorf("Expecting last ID to be %d; got %d", i, lastID2)
		}
		if lastIDErr2 != nil {
			t.Errorf("Expecting error to be nil; got %v", lastIDErr2)
		}
	}

	// Test 2: Unique Sentry DSN
	for i := 10; i < 20; i++ {
		storageDir := "."
		flowBulletinURL := "https://nifi.local/nifi-api/flow/bulletin-board"
		sentryDSN := fmt.Sprintf("https://wr23423we23423423423@sentry.local/%d", i)
		saveLastID(&storageDir, &flowBulletinURL, &sentryDSN, int64(i))

		// Should be i
		lastID2, lastIDErr2 := getLastID(&storageDir, &flowBulletinURL, &sentryDSN)
		if lastID2 != int64(i) {
			t.Errorf("Expecting last ID to be %d; got %d", i, lastID2)
		}
		if lastIDErr2 != nil {
			t.Errorf("Expecting error to be nil; got %v", lastIDErr2)
		}
	}
}

// Test whether parseNiFiIDString works as expected
func TestParseNiFiIDString(t *testing.T) {
	// Test 1: a number
	id := 231432
	newID, idErr1 := parseNiFiIDString(fmt.Sprintf("%d", id))
	if newID != int64(id) {
		t.Errorf("Expecting ID to be %d; got %d", id, newID)
	}
	if idErr1 != nil {
		t.Errorf("Expecting error to be ni; got %v", idErr1)
	}

	// Test 2: not a number
	_, idErr2 := parseNiFiIDString("notanumber")
	if idErr2 == nil {
		t.Errorf("Expecting error not to be nil; got %v", idErr2)
	}
}

func TestParseNiFiTimestampString(t *testing.T) {
	now := time.Now()
	nifiTimeFormat := strings.TrimSpace(strings.TrimLeft(nifiFormattedTimestampFormat, nifiMissingDateFormat))

	// Test 1: Correctly formatted time
	time1, timeErr1 := parseNiFiTimestampString(now.Format(nifiTimeFormat))
	if time1.Unix() != now.Unix() {
		t.Errorf("Expecting time to be %d; got %d", now.Unix(), time1.Unix())
	}
	if timeErr1 != nil {
		t.Errorf("Expecting error to be nil; got %v", timeErr1)
	}

	// Test 2: Empty timestamp
	_, timeErr2 := parseNiFiTimestampString("")
	if timeErr2 == nil {
		t.Errorf("Expecting error not to be nil; got %v", timeErr2)
	}
}
