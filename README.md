# Slipee

Slipee is a server, CLI tool and GO package for making static maps from a [slippy map](https://wiki.openstreetmap.org/wiki/Slippy_Map) type tile server.

## Motivation

Slipee fills the need for a server (microservice) to serve static map images to backend services in the [EIND](eind.no) tech stack. It is designed to do one thing, and one thing only:

> Provide static map images from a OSM-type tile server. The size and zoom and center lat/long coordinates of the image is adjustable.

## Installation

Download your binary from the [releases page](https://github.com/krilor/slipee/releases) and (otionally) put it in your path.

### Linux

```bash
wget -c https://github.com/krilor/slipee/releases/download/v0.0.2/slipee_0.0.2_Linux_x86_64.tar.gz -O slipee.tar.gz
tar --exclude='*.md' --exclude=LICENSE -xzf slipee.tar.gz
sudo mv slipee /usr/local/bin/slipee
rm slipee.tar.gz
```

### GO(lang)

```bash
go get -u github.com/krilor/slipee
```

## Example usage

Run the server using the command `slipee serve`. It will listen on port 8090.

Then visit the browser or curl an url like:

`http://localhost:7654/?zoom=16&width=500&height=500&lat=59.926181&long=10.775909`

Supported query args are:

* zoom
* width
* height
* lat
* long

These are the same as the ones mentioned in configuration below.

## Configuration

You can use command line flags or environment variables.

```
$ slipee help
USAGE:
  slipee <command> [flags]

COMMANDS:
  help           See this message
  serve          Serve slipee as a server

FLAGS:
  -address string
        the address to listen on
  -height int
        width in pixels (default 500)
  -label string
        the label to add to the image (default "Slipee | © OpenStreetMap contributors")
  -lat float
        latitude
  -long float
        longitude
  -port int
        port to listen on (default 7654)
  -tileserver string
        the tile server url with ${[xyz]} type variables (default "https://a.tile.openstreetmap.org/${z}/${x}/${y}.png")
  -width int
        width in pixels (default 500)
  -zoom int
        zoom level (default 16)

ENVIRONMENT VARIABLES:
  Flags parameters can also be specified as environment variables.
  Use uppercase flag name and the prefix 'SLIPEE_'.
  Example: tileserver -> SLIPEE_TILESERVER
  Command line flags have precedence over environment variables.
```

## TODOs

The following things needs to be done:

* Docker container
* Handle edge cases (like 0 zoom, where canvas is bigger than 256x256)
* Context
* Add more docs
* Add docs for Varnish cache recommendations

## Alternatives

Check out the the list over at [Open Street Map wiki](https://wiki.openstreetmap.org/wiki/Static_map_images). There are some pretty good alternatives.

## The name

_Slipee_ is taken from the [phonetic respelling](https://en.wikipedia.org/wiki/Pronunciation_respelling) of the word _slippy_, referring to a [slippy map](https://wiki.openstreetmap.org/wiki/Slippy_Map) map. It's prononced as _slip-ee_, aka. in [IPA](https://en.wikipedia.org/wiki/International_Phonetic_Alphabet), its / ˈslɪp i /


## Releasing with goreleaser

1. Check the current version tag using `git tag -l`
2. Create a new tag using `git tag -a v0.0.1 -m "some message"`
3. Push the tag using `git push origin v0.0.1`
4. Run goreleaser `goreleaser`
