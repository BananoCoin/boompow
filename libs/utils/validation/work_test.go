package validation

import (
	"testing"

	utils "github.com/bananocoin/boompow-next/libs/utils/testing"
)

func TestWorkValidation(t *testing.T) {
	// Valid result
	workResult := "205452237a9b01f4"
	hash := "3F93C5CD2E314FA16702189041E68E68C07B27961BF37F0B7705145BEFBA3AA3"

	// Check that we can access these items
	utils.AssertEqual(t, true, IsWorkValid(hash, 1, workResult))

	// Invalid result
	workResult = "205452237a9b01f4"
	hash = "3F93C5CD2E314FA16702189041E68E68C07B27961BF37F0B7705145BEFBA3AA3"

	// Check that we can access these items
	utils.AssertEqual(t, false, IsWorkValid(hash, 800, workResult))

	// Invalid result
	workResult = "205452237a9b01f4"
	hash = "F1C59E6C738BB82221E082910740BADC58301F8F32291E07CCC4CDBEEAD44348"

	// Check that we can access these items
	utils.AssertEqual(t, false, IsWorkValid(hash, 1, workResult))
}
