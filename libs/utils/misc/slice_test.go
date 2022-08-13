package misc

import (
	"testing"

	utils "github.com/bananocoin/boompow/libs/utils/testing"
)

func TestContains(t *testing.T) {
	arrayString := []string{"a", "b", "c"}
	arrayInt := []int{0, 1, 2}

	utils.AssertEqual(t, true, Contains(arrayString, "a"))
	utils.AssertEqual(t, false, Contains(arrayString, "d"))
	utils.AssertEqual(t, true, Contains(arrayInt, 0))
	utils.AssertEqual(t, false, Contains(arrayInt, 3))
}
