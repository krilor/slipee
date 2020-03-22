package query

import (
	"fmt"
	"net/url"
	"strconv"
)

// Package query provides convenience functions for working with request query parameters

// Int tries to get a query value
func Int(uv url.Values, key string, value int, min, max *int) (int, bool, error) {
	values, ok := uv[key]

	if !ok || (ok && len(values) < 1) {
		// query param is not found - return default
		return value, false, nil
	}

	queryValue, err := strconv.Atoi(values[0])

	if err != nil {
		return 0, ok, fmt.Errorf("%s is not an int", values[0])
	}

	// now we have the value - lets do some tests
	if min != nil {
		if queryValue < *min {
			return queryValue, ok, fmt.Errorf("%d is lower than %d", queryValue, *min)
		}
	}

	if max != nil {
		if queryValue > *max {
			return queryValue, ok, fmt.Errorf("%d is higher than %d", queryValue, *max)
		}
	}

	return queryValue, ok, nil
}

// Float64 tries to get a query value
func Float64(uv url.Values, key string, value float64, min, max *float64) (float64, bool, error) {
	values, ok := uv[key]

	if !ok || (ok && len(values) < 1) {
		// query param is not found - return default
		return value, false, nil
	}

	queryValue, err := strconv.ParseFloat(values[0], 64)

	if err != nil {
		return 0, ok, fmt.Errorf("%s is not a float", values[0])
	}

	// now we have the value - lets do some tests
	if min != nil {
		if queryValue < *min {
			return queryValue, ok, fmt.Errorf("%s is lower than %f", values[0], *min)
		}
	}

	if max != nil {
		if queryValue > *max {
			return queryValue, ok, fmt.Errorf("%s is higher than %f", values[0], *max)
		}
	}

	return queryValue, ok, nil
}

// Bool returns true if a query variable is present
func Bool(uv url.Values, key string) bool {
	_, present := uv[key]
	return present
}
