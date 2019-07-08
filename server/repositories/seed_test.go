package repositories_test

import (
	"chattingroom/server/repositories"
	"testing"
)

// go test -v seed_test.go
// go test -v -run TestSeed seed_test.go

func TestSeed(t *testing.T) {
	err := repositories.Seed()
	if err != nil {
		t.Fatalf(err.Error())
	}
}
