package tiles

import (
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/slippy"
	"math"
)

// Extent4326 returns a EPSG4326 extent for 't'. This is a utility method
// because I can't figure out how to do this using the available methods in
// go-spatial/geom package.
func Extent4326(t *slippy.Tile) *geom.Extent {

	return geom.NewExtent(
		[2]float64{tile2Lon(t.Z, t.X), tile2Lat(t.Z, t.Y+1)},
		[2]float64{tile2Lon(t.Z, t.X+1), tile2Lat(t.Z, t.Y)},
	)
}

func tile2Lon(zoom, x uint) float64 {
	return float64(x)/math.Exp2(float64(zoom))*360.0 - 180.0
}

func tile2Lat(zoom, y uint) float64 {
	var n float64 = math.Pi
	if y != 0 {
		n = math.Pi - 2.0*math.Pi*float64(y)/math.Exp2(float64(zoom))
	}

	return 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
}
