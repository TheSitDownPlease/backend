package main

import (
	"testing"
)

func TestStampToDate(t *testing.T) {

	timestamp := "1573876645782"
	excepted := "2019-11-16 11:57:25"
	result := stampToDate(timestamp)

	if excepted != result {
		t.Errorf("result:%s is not equal excepted %s", result, excepted)
	}

}
