package util

import "testing"

func AssertErrNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("Expect nil, but got: %+v", err)
	}
}

func AssertErrNotNil(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("Expect not nil, but got: %+v", err)
	}
}
