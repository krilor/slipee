package main

import (
	// image formats supported are commonly jpg or png
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"net/http"
	"strconv"

	"github.com/krilor/slipee/internal/tile"
	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
)

var s *tile.Server

func main() {
	s = tile.NewServer("http://localhost:9022/${z}/${x}/${y}.png")

	http.HandleFunc("/", static)
	http.ListenAndServe(":8090", nil)
}

func static(w http.ResponseWriter, req *http.Request) {

	width, _ := strconv.Atoi(req.URL.Query().Get("width"))
	height, _ := strconv.Atoi(req.URL.Query().Get("height"))
	zoom, _ := strconv.Atoi(req.URL.Query().Get("zoom"))
	lat, _ := strconv.ParseFloat(req.URL.Query().Get("lat"), 64)
	long, _ := strconv.ParseFloat(req.URL.Query().Get("long"), 64)

	img, err := s.StaticMap(width, height, zoom, lat, long)
	if err != nil {
		fmt.Fprintf(w, "an error occurred while patching image")
		return
	}

	label := "Slipee | © OpenStreetMap contributors"

	draw.DrawMask(
		img,
		image.Rectangle{image.Point{width - len(label)*8 - 16, height - 28}, image.Point{width, height}},
		&image.Uniform{color.RGBA{255, 255, 255, 255}}, // white
		image.Point{0, 0},
		&image.Uniform{color.Alpha{196}},
		image.Point{0, 0},
		draw.Over,
	)

	addLabel(img, width-(len(label))*8, height-8, label)

	enc := png.Encoder{
		CompressionLevel: png.BestSpeed,
	}
	err = enc.Encode(w, img)

	if err != nil {
		fmt.Fprintf(w, "an error occurred while serving image")
		return
	}

}

// addLabel is based on https://stackoverflow.com/a/38300583
func addLabel(img *image.RGBA, x, y int, label string) {
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
