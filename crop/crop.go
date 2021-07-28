// package coverage provides methods for cropping the geometry of Who's On First records.
package crop

import (
	"context"
	"fmt"
	"github.com/go-spatial/geom/slippy"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/clip"
	"github.com/paulmach/orb/geojson"
	"log"
)

func CropFeatureWithTile(ctx context.Context, body []byte, tile *slippy.Tile, grid slippy.Grid) ([]byte, error) {

	extent, ok := slippy.Extent(grid, tile)

	if !ok {
		return nil, fmt.Errorf("Failed to derive extent for tile '%v'", tile)
	}

	bounds := orb.Bound{
		Min: extent.Min(),
		Max: extent.Max(),
	}

	return CropFeatureWithBounds(ctx, body, bounds)
}

func CropFeatureWithBounds(ctx context.Context, body []byte, bounds orb.Bound) ([]byte, error) {

	f, err := geojson.UnmarshalFeature(body)

	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal feature, %w", err)
	}

	geom := f.Geometry

	clipped_geom := clip.Geometry(bounds, geom)

	if clipped_geom == nil {
		log.Println("CROP", bounds.Min.Lon(), bounds.Min.Lat(), bounds.Max.Lon(), bounds.Max.Lat())
		log.Println("FROM", geom.Bound())

		return nil, fmt.Errorf("Failed to derive clipped geometry")
	}

	f.Geometry = clipped_geom
	return f.MarshalJSON()
}
