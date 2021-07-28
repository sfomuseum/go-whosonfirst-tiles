// package coverage provides methods for cropping the geometry of Who's On First records.
package crop

import (
	"context"
	"fmt"
	"github.com/paulmach/orb/clip"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
)

func CropFeatureWithZXY(ctx context.Context, f *geojson.Feature, z uint, x uint, y uint) ([]byte, error) {

	zm := maptile.Zoom(uint32(z))
	tl := maptile.New(uint32(x), uint32(y), zm)

	bounds := tl.Bound()

	geom := f.Geometry
	clipped_geom := clip.Geometry(bounds, geom)

	if clipped_geom == nil {
		return nil, fmt.Errorf("Failed to derive clipped geometry")
	}

	f.Geometry = clipped_geom

	return f.MarshalJSON()
}
