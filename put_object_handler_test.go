package main

import (
	"fmt"
	"testing"
)

func TestParseDisposition(t *testing.T) {
	testCases := [][2]string{
		{`attachment; filename="test1.txt"`, "test1.txt"},
		{`inline; filename="image.jpg"`, "image.jpg"},
		{"inline", ""},
	}
	for _, testCase := range testCases {
		input := testCase[0]
		expect := testCase[1]
		result := ParseDisposition(input)
		t.Logf("INPUT = %s, EXPECT = %s, RESULT = %s\n", input, expect, result)
		fmt.Println(result)
		if result != expect {
			t.Fail()
		}
	}
}
