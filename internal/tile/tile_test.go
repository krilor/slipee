package tile

import (
	"fmt"
	"math"
	"testing"
)

// Used for float comarison
func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= 1e-13
}

func TestLatLongToWebMercator(t *testing.T) {

	var conversionTest = []struct {
		inLat     float64
		inLong    float64
		expectedX float64
		expectedY float64
	}{
		{41.85, -87.65, 65.67111111111112, 95.17492654697409}, // From https://developers.google.com/maps/documentation/javascript/coordinates
	}

	for _, test := range conversionTest {
		t.Run(fmt.Sprintf("%f,%f", test.inLat, test.inLong), func(t *testing.T) {
			x, y := latLongToWebMercator(test.inLat, test.inLong)
			// Almostequal is used since the test expected results have a finite presicion
			if !almostEqual(x, test.expectedX) || !almostEqual(y, test.expectedY) {
				t.Errorf("got %f,%f - want %f,%f", x, y, test.expectedX, test.expectedY)
			}
		})
	}
}

func TestMercatorToPixel(t *testing.T) {

	var conversionTest = []struct {
		in            float64
		zoom          int
		expectedTile  int
		expectedPixel int
	}{
		// From https://developers.google.com/maps/documentation/javascript/coordinates
		{65.67111111111112, 0, 0, 65},
		{95.17492654697409, 1, 0, 190},
		//{65.67111111111112, 95.17492654697409, 1, 131, 190},
		//{65.67111111111112, 95.17492654697409, 4, 1050, 1522},
		//{65.67111111111112, 95.17492654697409, 10, 67247, 97459},
	}

	for _, test := range conversionTest {
		t.Run(fmt.Sprintf("%f - %d", test.in, test.zoom), func(t *testing.T) {
			tile, pixel := mercatorToPixel(test.in, test.zoom)
			if tile != test.expectedTile || pixel != test.expectedPixel {
				t.Errorf("got %d,%d - want %d,%d", tile, pixel, test.expectedTile, test.expectedPixel)
			}
		})
	}
}

func TestServerSetURL(t *testing.T) {
	var urlTest = []struct {
		in       string
		expected string
	}{
		{"${z}", "%[1]d"},
		{"${Z}", "%[1]d"},
		{"${x}", "%[2]d"},
		{"${X}", "%[2]d"},
		{"${y}", "%[3]d"},
		{"${Y}", "%[3]d"},
		{"{z}", "%[1]d"},
		{"{Z}", "%[1]d"},
		{"{x}", "%[2]d"},
		{"{X}", "%[2]d"},
		{"{y}", "%[3]d"},
		{"{Y}", "%[3]d"},
	}

	for _, test := range urlTest {
		t.Run(test.in, func(t *testing.T) {
			s := Server{}
			s.setURL(test.in)

			if test.expected != s.server {
				t.Errorf("got %s - want %s", s.server, test.expected)
			}
		})
	}
}

func TestLatLimit(t *testing.T) {
	want := "85.051129"
	got := fmt.Sprintf("%f", latLimit)
	if got != want {
		t.Errorf("got %s - expected %s", got, want)
	}
}
