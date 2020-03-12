# Slipee

Slipee is a server, CLI tool and GO package for making static maps from a [slippy map](https://wiki.openstreetmap.org/wiki/Slippy_Map) type tile server.

## Motivation

Slipee fills the need for a server (microservice) to serve static map images to backend services in the [EIND](eind.no) tech stack. It is designed to do one thing, and one thing only:

> Provide static map images from a OSM-type tile server. The size and zoom and center lat/long coordinates of the image is adjustable.

## Example usage

Run the server using the command `slipee`. It will listen on port 8090.

Then visit the browser or curl an url like:

`http://localhost:8090/?zoom=16&width=500&height=500&lat=59.926181&long=10.775909`

## TODOs

The following things needs to be done:

* Add CLI params to specify tile server URL(s) and listen port
* Put attribution on the image
* Sanetize query params
* Add marker
* Handle edge cases (like 0 zoom, where canvas is bigger than 256x256)
* Context
* Add docs
* Add docs for Varnish cache recommendations
* Goreleaser
* Docker container

## Alternatives

Check out the the list over at [Open Street Map wiki](https://wiki.openstreetmap.org/wiki/Static_map_images). There are some pretty good alternatives.

## The name

_Slipee_ is taken from the [phonetic respelling](https://en.wikipedia.org/wiki/Pronunciation_respelling) of the word _slippy_, referring to a [slippy map](https://wiki.openstreetmap.org/wiki/Slippy_Map) map. Slipee is prononced as is, but in [IPA](https://en.wikipedia.org/wiki/International_Phonetic_Alphabet), its / ˈslɪp i /
