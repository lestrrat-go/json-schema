package validator

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNumericHelpers(t *testing.T) {
	t.Run("isNumeric", func(t *testing.T) {
		testCases := []struct {
			name string
			in   any
			want bool
		}{
			{name: "int", in: 5, want: true},
			{name: "int64", in: int64(5), want: true},
			{name: "uint64", in: uint64(5), want: true},
			{name: "float64", in: 5.5, want: true},
			{name: "json.Number int", in: json.Number("5"), want: true},
			{name: "json.Number float", in: json.Number("5.5"), want: true},
			{name: "string", in: "5", want: false},
			{name: "bool", in: true, want: false},
			{name: "nil", in: nil, want: false},
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				require.Equal(t, tc.want, isNumeric(tc.in))
			})
		}
	})

	t.Run("isJSONNumber", func(t *testing.T) {
		require.True(t, isJSONNumber(json.Number("5")))
		require.False(t, isJSONNumber("5"))
		require.False(t, isJSONNumber(5))
		require.False(t, isJSONNumber(nil))
	})

	t.Run("numericFloat", func(t *testing.T) {
		testCases := []struct {
			name    string
			in      any
			want    float64
			wantOK  bool
			wantErr bool
		}{
			{name: "int", in: 5, want: 5, wantOK: true},
			{name: "float64", in: 5.5, want: 5.5, wantOK: true},
			{name: "uint64", in: uint64(7), want: 7, wantOK: true},
			{name: "json.Number int", in: json.Number("5"), want: 5, wantOK: true},
			{name: "json.Number float", in: json.Number("5.5"), want: 5.5, wantOK: true},
			{name: "json.Number exponent", in: json.Number("1e2"), want: 100, wantOK: true},
			{name: "string", in: "5", wantOK: false},
			{name: "nil", in: nil, wantOK: false},
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				got, ok, err := numericFloat(tc.in)
				if tc.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				require.Equal(t, tc.wantOK, ok)
				if tc.wantOK {
					require.Equal(t, tc.want, got)
				}
			})
		}
	})

	t.Run("numericInt", func(t *testing.T) {
		testCases := []struct {
			name      string
			in        any
			want      int64
			wantOK    bool
			wantIsInt bool
			wantErr   bool
		}{
			{name: "int", in: 5, want: 5, wantOK: true, wantIsInt: true},
			{name: "int64", in: int64(5), want: 5, wantOK: true, wantIsInt: true},
			{name: "uint64", in: uint64(5), want: 5, wantOK: true, wantIsInt: true},
			{name: "negative", in: -3, want: -3, wantOK: true, wantIsInt: true},
			{name: "float integral", in: 5.0, want: 5, wantOK: true, wantIsInt: true},
			{name: "float fractional", in: 5.5, wantOK: true, wantIsInt: false},
			{name: "json.Number int", in: json.Number("5"), want: 5, wantOK: true, wantIsInt: true},
			{name: "json.Number 5.0", in: json.Number("5.0"), want: 5, wantOK: true, wantIsInt: true},
			{name: "json.Number 5.5", in: json.Number("5.5"), wantOK: true, wantIsInt: false},
			{name: "json.Number exponent", in: json.Number("1e2"), want: 100, wantOK: true, wantIsInt: true},
			{name: "json.Number negative", in: json.Number("-3"), want: -3, wantOK: true, wantIsInt: true},
			{name: "json.Number beyond 2^53", in: json.Number("9007199254740993"), want: 9007199254740993, wantOK: true, wantIsInt: true},
			{name: "json.Number out of int64 range", in: json.Number("99999999999999999999"), wantOK: true, wantErr: true},
			{name: "uint64 out of int64 range", in: uint64(math.MaxUint64), wantOK: true, wantErr: true},
			{name: "string", in: "5", wantOK: false},
			{name: "bool", in: true, wantOK: false},
			{name: "nil", in: nil, wantOK: false},
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				got, ok, isInt, err := numericInt(tc.in)
				if tc.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				require.Equal(t, tc.wantOK, ok)
				require.Equal(t, tc.wantIsInt, isInt)
				if tc.wantIsInt {
					require.Equal(t, tc.want, got)
				}
			})
		}
	})
}

// TestNumericValidatorsAcceptJSONNumber confirms the typed numeric validators
// treat a json.Number identically to the equivalent native value.
func TestNumericValidatorsAcceptJSONNumber(t *testing.T) {
	ctx := t.Context()

	t.Run("integer accepts json.Number integer", func(t *testing.T) {
		v := Integer().Minimum(0).Maximum(10).MustBuild()
		_, err := v.Validate(ctx, json.Number("5"))
		require.NoError(t, err)
	})
	t.Run("integer rejects fractional json.Number", func(t *testing.T) {
		v := Integer().MustBuild()
		_, err := v.Validate(ctx, json.Number("5.5"))
		require.Error(t, err)
	})
	t.Run("integer accepts integral 5.0 json.Number", func(t *testing.T) {
		v := Integer().MustBuild()
		_, err := v.Validate(ctx, json.Number("5.0"))
		require.NoError(t, err)
	})
	t.Run("number accepts json.Number float", func(t *testing.T) {
		v := Number().Minimum(0).MustBuild()
		_, err := v.Validate(ctx, json.Number("5.5"))
		require.NoError(t, err)
	})
	t.Run("string rejects json.Number under strict type", func(t *testing.T) {
		// strictStringType is set when a schema explicitly declares type: string.
		v := &stringValidator{strictStringType: true}
		_, err := v.Validate(ctx, json.Number("5"))
		require.Error(t, err)
	})
}
