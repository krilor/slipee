package query

import (
	"errors"
	"fmt"
	"net/url"
	"testing"
)

func TestInt(t *testing.T) {

	type in struct {
		v     url.Values
		key   string
		value int
		min   *int
		max   *int
	}
	type expect struct {
		value int
		ok    bool
		err   error
	}

	var min int = 0
	var max int = 100

	var intTest = []struct {
		in     in
		expect expect
	}{
		{in: in{url.Values(map[string][]string{"zoom": []string{"1"}}), "zoom", 0, nil, nil}, expect: expect{1, true, nil}},
		{in: in{url.Values(map[string][]string{"zoom": []string{"abc"}}), "zoom", 0, nil, nil}, expect: expect{0, true, errors.New("abc is not an int")}},
		{in: in{url.Values(map[string][]string{"zoom": []string{"2"}}), "zoom", 0, &min, &max}, expect: expect{2, true, nil}},
		{in: in{url.Values(map[string][]string{"zoom": []string{"0"}}), "zoom", 0, &min, &max}, expect: expect{0, true, nil}},
		{in: in{url.Values(map[string][]string{"zoom": []string{"-1"}}), "zoom", 0, &min, &max}, expect: expect{-1, true, errors.New("-1 is lower than 0")}},
		{in: in{url.Values(map[string][]string{"zoom": []string{"101"}}), "zoom", 0, &min, &max}, expect: expect{101, true, errors.New("101 is higher than 100")}},
	}

	for _, test := range intTest {
		t.Run(fmt.Sprintf("%+v", test.in), func(t *testing.T) {
			in := test.in
			value, ok, err := Int(in.v, in.key, in.value, in.min, in.max)

			// Almostequal is used since the test expected results have a finite presicion
			if value != test.expect.value {
				t.Errorf("value: got %d - want %d", value, test.expect.value)
			}
			if !equalError(err, test.expect.err) {
				t.Errorf("err: got '%v' - want '%v'", err, test.expect.err)
			}
			if ok != test.expect.ok {
				t.Errorf("ok: got %v - want %v", ok, test.expect.ok)
			}

		})
	}
}

func TestFloat64(t *testing.T) {

	type in struct {
		v     url.Values
		key   string
		value float64
		min   *float64
		max   *float64
	}
	type expect struct {
		value float64
		ok    bool
		err   error
	}

	var min float64 = 0.0
	var max float64 = 100.0

	var intTest = []struct {
		in     in
		expect expect
	}{
		{in: in{url.Values(map[string][]string{"zoom": []string{"1"}}), "zoom", 0, nil, nil}, expect: expect{1, true, nil}},
		{in: in{url.Values(map[string][]string{"zoom": []string{"abc"}}), "zoom", 0, nil, nil}, expect: expect{0, true, errors.New("abc is not a float")}},
		{in: in{url.Values(map[string][]string{"zoom": []string{"2.2"}}), "zoom", 0, &min, &max}, expect: expect{2.2, true, nil}},
		{in: in{url.Values(map[string][]string{"zoom": []string{"0"}}), "zoom", 0, &min, &max}, expect: expect{0, true, nil}},
		{in: in{url.Values(map[string][]string{"zoom": []string{"-1"}}), "zoom", 0, &min, &max}, expect: expect{-1, true, errors.New("-1 is lower than 0.000000")}},
		{in: in{url.Values(map[string][]string{"zoom": []string{"101"}}), "zoom", 0, &min, &max}, expect: expect{101, true, errors.New("101 is higher than 100.000000")}},
	}

	for _, test := range intTest {
		t.Run(fmt.Sprintf("%+v", test.in), func(t *testing.T) {
			in := test.in
			value, ok, err := Float64(in.v, in.key, in.value, in.min, in.max)

			// Almostequal is used since the test expected results have a finite presicion
			if value != test.expect.value {
				t.Errorf("value: got %f - want %f", value, test.expect.value)
			}
			if !equalError(err, test.expect.err) {
				t.Errorf("err: got '%v' - want '%v'", err, test.expect.err)
			}
			if ok != test.expect.ok {
				t.Errorf("ok: got %v - want %v", ok, test.expect.ok)
			}

		})
	}
}

// equalError is a utily to check if errors are equal
func equalError(a, b error) bool {
	return a == nil && b == nil || a != nil && b != nil && a.Error() == b.Error()
}
