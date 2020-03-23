package stitch

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"log"

	"github.com/pkg/errors"

	"github.com/krilor/slipee/internal/tile"
	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
)

// Package stitch implements the tile stitching operations
// It also controls the tile server access limitiations
// E.g. OSM tile servers should not be overloaded, and there is a maximum thread limit at 2.

// Request is a structure that holds all vars that make up a request
type request struct {
	width  int
	height int
	zoom   int
	lat    float64
	long   float64
	label  string
}

// Hash returns a hash string of the request, that can be used in caching type operations
func (r request) hash() string {
	hash := sha1.New()

	// errors are ignored on purpose. hash.Hash docs specifically state that the writer "never returns an error."
	binary.Write(hash, binary.LittleEndian, r.width)
	binary.Write(hash, binary.LittleEndian, r.height)
	binary.Write(hash, binary.LittleEndian, r.zoom)
	binary.Write(hash, binary.LittleEndian, r.lat)
	binary.Write(hash, binary.LittleEndian, r.long)

	hash.Write([]byte(r.label))

	return hex.EncodeToString(hash.Sum(nil))
}

func (r request) path() string {
	h := r.hash()
	return filepath.Join(h[:2], h[2:]+".png")

}

// Stitcher interface
type Stitcher interface {
	Stitch(width int, height int, zoom int, lat float64, long float64, label string) string
	Queue(width int, height int, zoom int, lat float64, long float64, label string) error
	StaticImage(width int, height int, zoom int, lat float64, long float64, label string) (string, error)
	StartWorker()
}

// New returns a new Stitcher for the tile server s.
// Size is the size of the queue buffer
func New(server *tile.Server, size int, cachePath string) Stitcher {
	s = stitch{
		server,
		make(chan request, size),
		cachePath,
	}

	return &s
}

// Get returns the stitcher singleton. Errors if it is not initialized
func Get() (Stitcher, error) {
	if s.queue == nil {
		return &s, errors.New("stitch singleton not initialized")
	}

	return &s, nil
}

// stitch is a struct that implements the stitcher interface
type stitch struct {
	server *tile.Server
	queue  chan request
	cache  string
}

// stitch is using a singleton pattern
var s stitch

// Get gets the path to a stitched static image
func (s *stitch) Stitch(width int, height int, zoom int, lat float64, long float64, label string) string {
	r := request{
		width,
		height,
		zoom,
		lat,
		long,
		label,
	}

	path := filepath.Join(s.cache, r.path())

	if _, err := os.Stat(path); err == nil {
		return path
	}

	// error is ignored on purpose
	s.Queue(width, height, zoom, lat, long, label)

	return ""

}

// Queue queues a request for later pickup
// Error is returned if channel is blocking (buffer is full)
func (s *stitch) Queue(width int, height int, zoom int, lat float64, long float64, label string) error {
	r := request{
		width,
		height,
		zoom,
		lat,
		long,
		label,
	}

	path := filepath.Join(s.cache, r.path())

	if _, err := os.Stat(path); err == nil {
		// image allready exists in file cache
		return nil
	}

	select {
	case s.queue <- r:
		return nil
	default:
		return errors.New("unable to queue")
	}

}

// StaticImage creates a static image
func (s *stitch) StaticImage(width int, height int, zoom int, lat float64, long float64, label string) (string, error) {

	r := request{
		width,
		height,
		zoom,
		lat,
		long,
		label,
	}
	path := filepath.Join(s.cache, r.path())

	os.MkdirAll(filepath.Dir(path), os.ModePerm) // TODO check err

	f, err := os.Create(path)
	if err != nil {
		return "", errors.Wrapf(err, "could not create file %s", path)
	}
	defer f.Close()

	img, err := s.server.StaticMap(width, height, zoom, lat, long)
	if err != nil {
		return "", errors.Wrap(err, "an error occurred while getting staticmap")
	}

	// to embed another PNG marker, use the following command in your terminal
	// cat some-marker-24.png | base64 -w 0 | xclip -sel clip
	data, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAABgAAAAYCAYAAADgdz34AAAABmJLR0QA/wD/AP+gvaeTAAABJElEQVRIieXUPUoDQRjG8R8qaKcgBsHKGPAAFoK2HkE9Qu5grXcQWysjaBsrrVIavYFFWkGNFpoiWuwGlt3ZuJtNIz7wws687/yf+dgZ/oNqOMEDPuLo4jjOVdIB+vjOiT72q8CHY+CjGE5iUvtl5ul4w0oINJtjcIS9VN8rznGPBhYSuXl84q7oCh5TM3zBeiJfjw2TNd2icHhPDT4N1JzJHnhGMzkGXwXq0n2DHFZQHdktqifyG7Jb1AmB5nIMbrGTaC+J9vgybh9iMTCmsBqK3YFkbJYxgHYJeLssHLZKrGJ7EgNoFYC3JoXDqugPGvdErFUxgOYYg2ZV+EgXAfjVtOBE9+ApAe9heZoGsCt6DgbxdyHlPdch9fCMG1yXmtqf1g/2CJPvQAzABQAAAABJRU5ErkJggg==")
	marker, _ := png.Decode(bytes.NewReader(data))

	addLabel(img, label)
	addMarker(img, marker)

	enc := png.Encoder{
		CompressionLevel: png.BestSpeed,
	}

	err = enc.Encode(f, img)

	if err != nil {
		return "", errors.Wrapf(err, "could not encode image")
	}

	return path, nil
}

// StartWorker spins of a goroutine that has a worker creating images off the queue
func (s *stitch) StartWorker() {
	go func(s *stitch) {
		for r := range s.queue {
			_, err := s.StaticImage(r.width, r.height, r.zoom, r.lat, r.long, r.label)

			if err != nil {
				log.Printf("could not create staticimage for r: %+v due to: %s", r, err)
			} else {
				log.Printf("staticimage created for r: %+v", r)
			}

			time.Sleep(time.Second)
		}

	}(s)

}

// addLabel is based on https://stackoverflow.com/a/38300583
func addLabel(img *image.RGBA, label string) {

	b := img.Bounds()
	width := b.Dx()
	height := b.Dy()

	// adds white area for label
	draw.DrawMask(
		img,
		image.Rectangle{image.Point{width - len(label)*8 - 16, height - 28}, image.Point{width, height}},
		&image.Uniform{color.RGBA{255, 255, 255, 255}}, // white
		image.Point{0, 0},
		&image.Uniform{color.Alpha{196}},
		image.Point{0, 0},
		draw.Over,
	)

	x := width - (len(label))*8
	y := height - 8

	col := color.RGBA{0, 0, 0, 255}

	// TODO make this point stuff more understandable
	point := fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: inconsolata.Regular8x16,
		Dot:  point,
	}
	d.DrawString(label)
}

func addMarker(img *image.RGBA, marker image.Image) {

	b := img.Bounds()
	width := b.Dx()
	height := b.Dy()

	draw.Draw(
		img,
		image.Rectangle{image.Point{width/2 - 12, height/2 - 12}, image.Point{width, height}},
		marker,
		image.Point{0, 0},
		draw.Over,
	)
}
