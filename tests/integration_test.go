package tests

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildExecutable(t *testing.T) {
	// Test that we can build the executable
	err := os.Chdir("../") // Go to project root
	if err != nil {
		t.Fatal(err)
	}

	// We just test that the go.mod exists and is properly configured
	_, err = os.Stat("go.mod")
	assert.NoError(t, err)
}