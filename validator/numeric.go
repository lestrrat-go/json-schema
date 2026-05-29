package validator

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
)

// The helpers in this file are the single place that decides whether an
// incoming value is a JSON number and what its value is. They accept BOTH
// native Go numeric kinds (int*, uint*, float*) — as produced by json.Unmarshal,
// struct fields, or builder-supplied literals — AND json.Number, as produced by
// a decoder configured with UseNumber (see ValidateJSON). json.Number is a named
// string type, so reflect.ValueOf(json.Number("1")).Kind() reports
// reflect.String; every numeric site must therefore consult these helpers rather
// than switching on reflect.Kind directly, and the string validator must exclude
// json.Number (see isJSONNumber).

// isNumeric reports whether v is a JSON number value: any native Go numeric kind
// OR a json.Number. A plain string is never considered numeric.
func isNumeric(v any) bool {
	if _, ok := v.(json.Number); ok {
		return true
	}
	switch reflect.ValueOf(v).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// isJSONNumber reports whether v is specifically a json.Number. The string
// validator uses this to exclude numbers that would otherwise report
// reflect.String kind and be validated as strings.
func isJSONNumber(v any) bool {
	_, ok := v.(json.Number)
	return ok
}

// numericFloat converts a numeric value (native kind or json.Number) to float64.
// ok is false when v is not numeric at all. err is non-nil only when v is a
// json.Number whose text cannot be parsed as a float (it should not happen for
// decoder output, but is surfaced rather than silently swallowed).
func numericFloat(v any) (float64, bool, error) {
	if n, ok := v.(json.Number); ok {
		f, err := n.Float64()
		if err != nil {
			return 0, true, err
		}
		return f, true, nil
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), true, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(rv.Uint()), true, nil
	case reflect.Float32, reflect.Float64:
		return rv.Float(), true, nil
	default:
		return 0, false, nil
	}
}

// numericInt converts a numeric value to an int64, preserving precision for
// json.Number via Int64() (so integers in the 2^53..2^63 range that float64
// would round are exact). The three return signals are:
//   - ok=false:              v is not numeric at all (caller reports a type error)
//   - ok=true, isInt=false:  numeric but not an integer (e.g. 5.5); n is unset
//   - ok=true, isInt=true:   n holds the integer value
//
// A non-nil err means v is an integer-valued number outside the int64 range
// (±9.2e18); such values cannot be validated as integers and are reported
// rather than silently truncated.
func numericInt(v any) (int64, bool, bool, error) {
	if num, ok := v.(json.Number); ok {
		if i, err := num.Int64(); err == nil {
			return i, true, true, nil
		}
		// Int64 rejects exponent forms (e.g. "1e2") and out-of-range values.
		// Fall back to Float64 to tell an integral value apart from a fractional
		// one, and from one that simply does not fit in int64.
		f, err := num.Float64()
		if err != nil {
			return 0, true, false, err
		}
		return integralFloatToInt64(f, num.String())
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), true, true, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u := rv.Uint()
		if u > math.MaxInt64 {
			return 0, true, false, fmt.Errorf("integer value %d out of int64 range", u)
		}
		return int64(u), true, true, nil
	case reflect.Float32, reflect.Float64:
		return integralFloatToInt64(rv.Float(), "")
	default:
		return 0, false, false, nil
	}
}

// integralFloatToInt64 reports whether f is an integer value and, if so, returns
// it as int64. text, when non-empty, is the original json.Number text used for a
// precise out-of-range error message.
func integralFloatToInt64(f float64, text string) (int64, bool, bool, error) {
	if f != math.Trunc(f) {
		return 0, true, false, nil // fractional: numeric but not an integer
	}
	// float64(math.MaxInt64) rounds up to 2^63, so the upper bound is exclusive.
	if f < math.MinInt64 || f >= math.MaxInt64 {
		if text == "" {
			return 0, true, false, fmt.Errorf("integer value %v out of int64 range", f)
		}
		return 0, true, false, fmt.Errorf("integer value %s out of int64 range", text)
	}
	return int64(f), true, true, nil
}
