package main

import (
	// image formats supported are commonly jpg or png

	"flag"
	"fmt"
	_ "image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/krilor/slipee/internal/env"
	"github.com/krilor/slipee/internal/query"
	"github.com/krilor/slipee/internal/stitch"
	"github.com/krilor/slipee/internal/tile"
)

var s stitch.Stitcher

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
	pronto     bool
	queue      int
	cache      string
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
	flag.BoolVar(&config.pronto, "pronto", env.Bool("SLIPEE_PRONTO", false), "if clients are allowed to buypass queue and ask for static images promtly")
	flag.IntVar(&config.queue, "queue", env.Int("SLIPEE_QUEUE", 1000), "queue size")
	flag.StringVar(&config.cache, "cache", env.String("SLIPEE_CACHE", "./slipee_cache"), "directory for cached maps")

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

	// TODO - can we make things work without globals?
	s = stitch.New(tile.NewServer(config.tileserver), config.queue, config.cache)
	s.StartWorker()

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

	// TODO - min/max for lat/long
	lat, _, err := query.Float64(uv, "lat", config.lat, nil, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("bad lat value: %s", err), 400)
		return
	}

	long, _, err := query.Float64(uv, "long", config.lat, nil, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("bad lat value: %s", err), 400)
		return
	}

	pronto := query.Bool(uv, "pronto") && config.pronto

	if req.Method == http.MethodPost {
		err := s.Queue(width, height, zoom, lat, long, config.label)
		if err != nil {
			http.Error(w, "could not queue", 500)
			return
		}
	}
	if req.Method != http.MethodGet {
		http.Error(w, "http method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var path string
	if pronto {
		path, err = s.StaticImage(width, height, zoom, lat, long, config.label)
		if err != nil {
			log.Println(err)
			http.Error(w, "could not get static image", 500)
			return
		}
	} else {
		path = s.Stitch(width, height, zoom, lat, long, config.label)
	}

	if path == "" {
		w.WriteHeader(http.StatusAccepted)
		return // TODO - return a transparent image
	}

	b, err := ioutil.ReadFile(path)

	if err != nil {
		http.Error(w, "could not read static image", 500)
		return
	}

	_, err = w.Write(b)

	if err != nil {
		http.Error(w, "could not write image back", 500)
		return
	}

	return
}

// Usage prints the cli usage
func usage() {
	flag.Usage()
}
