package validator_test

import (
	"testing"

	"github.com/lestrrat-go/json-schema/validator"
)

func TestObjectConstrainctSanity(t *testing.T) {
	testcases := makeSanityTestCases()
	for _, tc := range testcases {
		switch tc.Name {
		case "Empty Map", "Empty Object":
		default:
			tc.Error = true
		}
	}

	var c validator.ObjectValidator
	for _, tc := range testcases {
		t.Run(tc.Name, makeSanityTestFunc(tc, &c))
	}
}
