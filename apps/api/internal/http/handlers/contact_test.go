package handlers

import "testing"

func TestMaskPhoneKeepsOnlyTheEnds(t *testing.T) {
	cases := map[string]string{
		"081234567890": "08********90",
		"1234":         "****",
		"":             "",
		"12345":        "12*45",
	}
	for in, want := range cases {
		if got := maskPhone(in); got != want {
			t.Errorf("maskPhone(%q) = %q, want %q", in, got, want)
		}
	}
}
