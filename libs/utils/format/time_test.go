package format

import (
	"testing"
	"time"

	utils "github.com/bananocoin/boompow-next/libs/utils/testing"
)

func TestFormatTime(t *testing.T) {
	time := time.Unix(1659715327, 0)
	formatted := GenerateISOString(time.UTC())

	utils.AssertEqual(t, "2022-08-05T16:02:07Z", formatted)
}
