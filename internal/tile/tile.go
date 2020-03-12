package tile

import (
	"fmt"
	"image"
	"image/draw"

	"math"
	"net/http"
	"regexp"

	"github.com/pkg/errors"
)

// latLimit is the upper/lower latitude limit for web map
var latLimit float64 = (math.Atan(math.Sinh(math.Pi)) / (2.0 * math.Pi)) * 360.0

// find takes a lat, long and zoom and returns a tile and a image.Point for the coordinate that represents the position.
func find(lat, long float64, zoom int) (int, int, image.Point) {

	pp := image.Point{}
	var x int
	var y int

	mx, my := latLongToWebMercator(lat, long)

	x, pp.X = mercatorToPixel(mx, zoom)
	y, pp.Y = mercatorToPixel(my, zoom)

	return x, y, pp
}

// Server is a servever to get tiles from
// https://wiki.openstreetmap.org/wiki/tile_servers
type Server struct {
	server string
}

// NewServer returns a new Server for the given URL
// url takes the format
// https://a.tile.openstreetmap.org/${z}/${x}/${y}.png
func NewServer(url string) *Server {
	s := Server{}
	s.setURL(url)
	return &s
}

// setUrl is used to set the Server url
// The primary purpose of this func is to be able to accept the commonly used ${X}/${x} variables used in strings.
func (s *Server) setURL(url string) {

	replacements := map[*regexp.Regexp]string{
		regexp.MustCompile(`\$?{[zZ]}`): "%[1]d",
		regexp.MustCompile(`\$?{[xX]}`): "%[2]d",
		regexp.MustCompile(`\$?{[yY]}`): "%[3]d",
	}

	s.server = url
	for re, replacement := range replacements {
		s.server = re.ReplaceAllString(s.server, replacement)
	}
	// TODO - verify URL format
	return
}

// Get returns a image from the tile server
func (s Server) Get(x, y, zoom int) (image.Image, error) {
	if s.server == "" {
		return nil, errors.New("server URL is missing - use NewServer to initialize the server")
	}
	url := fmt.Sprintf(s.server, zoom, x, y)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	req.Header.Add("User-Agent", "Slipee/0.0 (+https://github.com/krilor/slipee)")
	// TODO - automate version in UA
	// TODO HTTP Referrer header

	res, err := client.Do(req)

	if err != nil {
		fmt.Println("in err")
		return nil, errors.Wrapf(err, "could not get tile %d/%d/%d", zoom, x, y)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("got status code %d for %s", res.StatusCode, url)
	}

	img, _, err := image.Decode(res.Body)

	if err != nil {
		return nil, errors.Wrapf(err, "could not decode tile %d/%d/%d", zoom, x, y)
	}

	return img, nil
}

// Find returs a tile image based on latitude and longitude.
// The image.Point returned is the pixel coordinate of the lat/long position.
// Integers returned are the x/y tile numbers
func (s Server) Find(lat, long float64, zoom int) (image.Image, image.Point, int, int, error) {
	x, y, p := find(lat, long, zoom)
	img, err := s.Get(x, y, zoom)

	if err != nil {
		return img, p, 0, 0, errors.Wrapf(err, "could not get tile for %f,%f-%d", lat, long, zoom)
	}

	return img, p, x, y, nil
}

// StaticMap patches together a image.Image of widht*height with lat and long in center. Zoom is the zoom level.
func (s Server) StaticMap(width, height, zoom int, lat, long float64) (*image.RGBA, error) {
	tileX, tileY, p := find(lat, long, zoom)

	// Now we need to figure out a few things about the image
	center := image.Point{int(width / 2), int(height / 2)}
	offset := image.Point{-(256 - (center.X-p.X)%256), -(256 - (center.Y-p.Y)%256)}
	nX := (width-offset.X)/256 + 1
	nY := (height-offset.Y)/256 + 1
	startX := tileX - (center.X-offset.X)/256
	startY := tileY - (center.Y-offset.Y)/256

	static := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{width, height}})

	for x := 0; x < nX; x++ {
		for y := 0; y < nY; y++ {
			img, err := s.Get(startX+x, startY+y, zoom)
			if err != nil {
				return nil, errors.Wrap(err, "could not get tile in loop")
			}

			draw.Draw(
				static,
				image.Rectangle{image.Point{256*x + offset.X, 256*y + offset.Y}, image.Point{256*(x+1) + offset.X, 256*(y+1) + offset.Y}},
				img,
				image.Point{0, 0},
				draw.Src,
			)
		}
	}

	return static, nil
}

// latLongToWebMercator converts from lat and long to Web Mercator
// https://en.wikipedia.org/wiki/Web_Mercator_projection
// https://developers.google.com/maps/documentation/javascript/coordinates
func latLongToWebMercator(lat, long float64) (float64, float64) {

	x := ((long + 180) / 360) * 256.0

	// y is implemented based on the formula here: https://en.wikipedia.org/wiki/Web_Mercator_projection
	latRadians := lat * math.Pi / 180
	y := (256 / (2.0 * math.Pi)) * (math.Pi - math.Log(math.Tan((math.Pi/4.0)+(latRadians/2.0))))

	return x, y
}

// mercatorToPixel takes a lat or long mercator and a zoom level and returns the tile number and pixel within that tile
// To get the total pixel position, use
//
//    absolutePixel = tile * 256 + pixel
//
func mercatorToPixel(m float64, zoom int) (int, int) {
	absolutePixel := int(float64(int(1)<<zoom) * m)
	return absolutePixel / 256, absolutePixel & 255
}
