package main

import (
	"flag"
	"github.com/go-spatial/geom/slippy"
	"github.com/sfomuseum/go-whosonfirst-tiles"
	"log"
)

func main() {

	z := flag.Int("z", 15, "...")
	x := flag.Int("x", 5244, "...")
	y := flag.Int("y", 12683, "...")

	flag.Parse()

	t := slippy.NewTile(uint(*z), uint(*x), uint(*y))

	log.Println(tiles.Extent4326(t))
}
