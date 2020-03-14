package main

import (
	// image formats supported are commonly jpg or png

	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/krilor/slipee/internal/env"
	"github.com/krilor/slipee/internal/query"
	"github.com/krilor/slipee/internal/tile"
	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
)

var s *tile.Server

// config holds cli variables
var config struct {
	lat        float64
	long       float64
	width      int
	height     int
	zoom       int
	tileserver string
	address    string
	port       int
	label      string
}

func init() {
	flag.Float64Var(&config.lat, "lat", env.Float64("SLIPEE_LAT", 0.0), "latitude")
	flag.Float64Var(&config.long, "long", env.Float64("SLIPEE_LONG", 0.0), "longitude")
	flag.IntVar(&config.width, "width", env.Int("SLIPEE_WIDTH", 500), "width in pixels")
	flag.IntVar(&config.height, "height", env.Int("SLIPEE_HEIGHT", 500), "width in pixels")
	flag.IntVar(&config.zoom, "zoom", env.Int("SLIPEE_ZOOM", 16), "zoom level")
	flag.StringVar(&config.address, "address", env.String("SLIPEE_ADDRESS", ""), "the address to listen on")
	flag.IntVar(&config.port, "port", env.Int("SLIPEE_PORT", 7654), "port to listen on")
	flag.StringVar(&config.tileserver, "tileserver", env.String("SLIPEE_TILESERVER", "https://a.tile.openstreetmap.org/${z}/${x}/${y}.png"), "the tile server url with ${[xyz]} type variables")
	flag.StringVar(&config.label, "label", env.String("SLIPEE_LABEL", "Slipee | © OpenStreetMap contributors"), "the label to add to the image")

	flag.Usage = func() {
		fmt.Println(`USAGE:
  slipee <command> [flags]

COMMANDS:
  help           See this message
  serve          Serve slipee as a server`)
		fmt.Fprint(flag.CommandLine.Output(), "\nFLAGS:\n")
		flag.PrintDefaults()

		fmt.Println(`
ENVIRONMENT VARIABLES:
  Flags parameters can also be specified as environment variables.
  Use uppercase flag name and the prefix 'SLIPEE_'.
  Example: tileserver -> SLIPEE_TILESERVER
  Command line flags have precedence over environment variables.`)
	}
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("No command specified")
		usage()
		os.Exit(1)
	}

	// Picking out the command like this is done because flag.Parse() stops at the first non-flag argument.
	// This allows us to have the CLI format slipee <command> [flags]
	command := os.Args[1]
	if len(os.Args) > 2 {
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}

	flag.Parse()

	switch command {
	case "serve":
		log.Println("starting server...")
		serve()
	case "help":
		usage()
		os.Exit(0)
	default:
		fmt.Printf("Unrecognized command '%s' specified\n", command)
		fmt.Println("Use 'slipee help' to see available commands")
		os.Exit(1)
	}
}

// serve handles the serve command
func serve() {

	s = tile.NewServer(config.tileserver)

	http.HandleFunc("/", static)
	log.Printf("server config: %+v\n", config)
	log.Printf("listening on %s:%d\n", config.address, config.port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", config.address, config.port), nil))

}

func static(w http.ResponseWriter, req *http.Request) {

	uv, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		http.Error(w, fmt.Sprintf("bad request %s", err), 400)
		return
	}

	// size
	minSize := 0
	maxSize := 2000

	width, _, err := query.Int(uv, "width", config.width, &minSize, &maxSize)
	if err != nil {
		http.Error(w, fmt.Sprintf("bad width value: %s", err), 400)
		return
	}

	height, _, err := query.Int(uv, "height", config.height, &minSize, &maxSize)
	if err != nil {
		http.Error(w, fmt.Sprintf("bad height value: %s", err), 400)
		return
	}

	// zoom
	minZoom := 0
	maxZoom := 23
	zoom, _, err := query.Int(uv, "zoom", config.zoom, &minZoom, &maxZoom)
	if err != nil {
		http.Error(w, fmt.Sprintf("bad zoom value: %s", err), 400)
		return
	}

	lat, _ := strconv.ParseFloat(req.URL.Query().Get("lat"), 64)
	long, _ := strconv.ParseFloat(req.URL.Query().Get("long"), 64)

	img, err := s.StaticMap(width, height, zoom, lat, long)
	if err != nil {
		log.Println(err)
		http.Error(w, "an error occurred while getting image", 500)
		return
	}

	// to embed another PNG marker, use the following command in your terminal
	// cat some-marker-24.png | base64 -w 0 | xclip -sel clip
	data, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAABgAAAAYCAYAAADgdz34AAAABmJLR0QA/wD/AP+gvaeTAAABJElEQVRIieXUPUoDQRjG8R8qaKcgBsHKGPAAFoK2HkE9Qu5grXcQWysjaBsrrVIavYFFWkGNFpoiWuwGlt3ZuJtNIz7wws687/yf+dgZ/oNqOMEDPuLo4jjOVdIB+vjOiT72q8CHY+CjGE5iUvtl5ul4w0oINJtjcIS9VN8rznGPBhYSuXl84q7oCh5TM3zBeiJfjw2TNd2icHhPDT4N1JzJHnhGMzkGXwXq0n2DHFZQHdktqifyG7Jb1AmB5nIMbrGTaC+J9vgybh9iMTCmsBqK3YFkbJYxgHYJeLssHLZKrGJ7EgNoFYC3JoXDqugPGvdErFUxgOYYg2ZV+EgXAfjVtOBE9+ApAe9heZoGsCt6DgbxdyHlPdch9fCMG1yXmtqf1g/2CJPvQAzABQAAAABJRU5ErkJggg==")
	marker, _ := png.Decode(bytes.NewReader(data))

	addLabel(img, config.label)
	addMarker(img, marker)

	enc := png.Encoder{
		CompressionLevel: png.BestSpeed,
	}

	err = enc.Encode(w, img)

	if err != nil {
		log.Println(err)
		http.Error(w, "an error occurred while serving image", 500)
		return
	}

}

// addLabel is based on https://stackoverflow.com/a/38300583
func addLabel(img *image.RGBA, label string) {

	b := img.Bounds()
	width := b.Dx()
	height := b.Dy()

	// adds white area for label
	draw.DrawMask(
		img,
		image.Rectangle{image.Point{width - len(config.label)*8 - 16, height - 28}, image.Point{width, height}},
		&image.Uniform{color.RGBA{255, 255, 255, 255}}, // white
		image.Point{0, 0},
		&image.Uniform{color.Alpha{196}},
		image.Point{0, 0},
		draw.Over,
	)

	x := width - (len(config.label))*8
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

// Usage prints the cli usage
func usage() {
	flag.Usage()
}
