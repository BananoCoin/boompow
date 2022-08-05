package format

import (
	"testing"
	"time"

	utils "github.com/bananocoin/boompow-next/libs/utils/testing"
)

func TestFormatTime(t *testing.T) {
	formatted := GenerateISOString(time.Unix(1659715327, 0))

	utils.AssertEqual(t, "2022-08-05T12:02:07-04:00", formatted)
}
