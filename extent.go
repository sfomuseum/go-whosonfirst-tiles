package tiles

import (
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/slippy"
	"math"
)

func Tile2Lon(zoom, x uint) float64 {
	return float64(x)/math.Exp2(float64(zoom))*360.0 - 180.0
}

func Tile2Lat(zoom, y uint) float64 {
	var n float64 = math.Pi
	if y != 0 {
		n = math.Pi - 2.0*math.Pi*float64(y)/math.Exp2(float64(zoom))
	}

	return 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
}

func Extent4326(t *slippy.Tile) *geom.Extent {

	return geom.NewExtent(
		[2]float64{Tile2Lon(t.Z, t.X), Tile2Lat(t.Z, t.Y+1)},
		[2]float64{Tile2Lon(t.Z, t.X+1), Tile2Lat(t.Z, t.Y)},
	)
}
