package validation

import (
	"encoding/hex"
	"testing"

	utils "github.com/bananocoin/boompow-next/libs/utils/testing"
)

func TestReverseOrder(t *testing.T) {
	str := []byte{1, 2, 3}

	utils.AssertEqual(t, []byte{3, 2, 1}, Reversed(str))
}

func TestReverseUnordered(t *testing.T) {
	str := []byte{1, 2, 1, 3, 1}

	utils.AssertEqual(t, []byte{1, 3, 1, 2, 1}, Reversed(str))
}

func TestAddressToPub(t *testing.T) {
	pub, _ := AddressToPub("ban_3t6k35gi95xu6tergt6p69ck76ogmitsa8mnijtpxm9fkcm736xtoncuohr3")

	utils.AssertEqual(t, "e89208dd038fbb269987689621d52292ae9c35941a7484756ecced92a65093ba", hex.EncodeToString(pub))
}

func TestValidateAddress(t *testing.T) {
	// Valid
	valid := "ban_1zyb1s96twbtycqwgh1o6wsnpsksgdoohokikgjqjaz63pxnju457pz8tm3r"
	utils.AssertEqual(t, true, ValidateAddress(valid))
	// Invalid
	invalid := "ban_1zyb1s96twbtycqwgh1o6wsnpsksgdoohokikgjqjaz63pxnju457pz8tm3ra"
	utils.AssertEqual(t, false, ValidateAddress(invalid))
	invalid = "ban_1zyb1s96twbtycqwgh1o6wsnpsksgdoohokikgjqjaz63pxnju457pz8tm3rb"
	utils.AssertEqual(t, false, ValidateAddress(invalid))
	invalid = "nano_1zyb1s96twbtycqwgh1o6wsnpsksgdoohokikgjqjaz63pxnju457pz8tm3r"
	utils.AssertEqual(t, false, ValidateAddress(invalid))
}
