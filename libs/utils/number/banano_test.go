package number

import "testing"

func TestRawToBanano(t *testing.T) {
	// 1
	raw := "100000000000000000000000000000"
	expected := 1.0
	converted, _ := RawToBanano(raw, true)
	if converted != expected {
		t.Errorf("Expected %f but got %f", expected, converted)
	}
	// 1.01
	raw = "101000000000000000000000000000"
	expected = 1.01
	converted, _ = RawToBanano(raw, true)
	if converted != expected {
		t.Errorf("Expected %f but got %f", expected, converted)
	}
	// 1.019
	raw = "101900000000000000000000000000"
	expected = 1.01
	converted, _ = RawToBanano(raw, true)
	if converted != expected {
		t.Errorf("Expected %f but got %f", expected, converted)
	}
	// 100000
	raw = "10000000000000000000000000000000000"
	expected = 100000
	converted, _ = RawToBanano(raw, true)
	if converted != expected {
		t.Errorf("Expected %f but got %f", expected, converted)
	}
	// Error
	raw = "1234NotANumber"
	expected = 1234.123456
	_, err := RawToBanano(raw, true)
	if err == nil {
		t.Errorf("Expected error converting %s but didn't get one", raw)
	}
}

func TestBananoToRaw(t *testing.T) {
	// 1
	expected := "100000000000000000000000000000"
	amount := 1.0
	converted := BananoToRaw(amount)
	if converted != expected {
		t.Errorf("Expected %s but got %s", expected, converted)
	}
	// 1.01
	expected = "101000000000000000000000000000"
	amount = 1.01
	converted = BananoToRaw(amount)
	if converted != expected {
		t.Errorf("Expected %s but got %s", expected, converted)
	}
	// 100000
	expected = "10000000000000000000000000000000000"
	amount = 100000
	converted = BananoToRaw(amount)
	if converted != expected {
		t.Errorf("Expected %s but got %s", expected, converted)
	}
}

// 100000000000000000000000000000
// 10000000000000000000000000000000
