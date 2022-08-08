package number

import (
	"errors"
	"fmt"
	"math"
	"math/big"
)

const rawPerBananoStr = "100000000000000000000000000000"

var rawPerBanano, _ = new(big.Float).SetString(rawPerBananoStr)

const bananoPrecision = 100 // 0.01 BANANO precision

// Raw to Big - converts raw amount to a big.Int
func RawToBigInt(raw string) (*big.Int, error) {
	rawBig, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Unable to convert %s to big int", raw))
	}
	return rawBig, nil
}

// RawToBanano - Converts Raw amount to usable Banano amount
func RawToBanano(raw string, truncate bool) (float64, error) {
	rawBig, ok := new(big.Float).SetString(raw)
	if !ok {
		err := errors.New(fmt.Sprintf("Unable to convert %s to int", raw))
		return -1, err
	}
	asBanano := rawBig.Quo(rawBig, rawPerBanano)
	f, _ := asBanano.Float64()
	if !truncate {
		return f, nil
	}

	return math.Trunc(f/0.01) * 0.01, nil
}

// BananoToRaw - Converts Banano amount to Raw amount
func BananoToRaw(banano float64) string {
	bananoInt := int(banano * 100)
	bananoRaw, _ := new(big.Int).SetString("1000000000000000000000000000", 10)

	res := bananoRaw.Mul(bananoRaw, big.NewInt(int64(bananoInt)))

	return fmt.Sprintf("%d", res)
}
