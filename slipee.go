package main

import (
	// image formats supported are commonly jpg or png
	"fmt"
	_ "image/jpeg"
	"image/png"
	"net/http"
	"strconv"

	"github.com/krilor/slipee/internal/tile"
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

	enc := png.Encoder{
		CompressionLevel: png.BestSpeed,
	}
	err = enc.Encode(w, img)

	if err != nil {
		fmt.Fprintf(w, "an error occurred while serving image")
		return
	}

}
