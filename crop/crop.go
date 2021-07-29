// package coverage provides methods for cropping the geometry of Who's On First records.
package crop

import (
	"context"
	"fmt"
	_ "github.com/go-spatial/geom/slippy"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/clip"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"log"
)

// CropFeatureWithTile will crop the geometry of a GeoJSON Feature defined by 'body' to the extent of 'tile'.
func CropFeatureWithTile(ctx context.Context, body []byte, tile maptile.Tile) ([]byte, error) {

	bounds := tile.Bound()
	return CropFeatureWithBounds(ctx, body, bounds)
}

// CropFeatureWithTile will crop the geometry of a GeoJSON Feature defined by 'body' to the extent of 'bounds'.
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
