// package coverage provides methods for cropping the geometry of Who's On First records.
package crop

import (
	"context"
	"fmt"
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/planar/clip"
	"github.com/paulmach/orb"
	// "github.com/paulmach/orb/clip"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/sfomuseum/go-whosonfirst-tiles/geometry"
	_ "log"
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

	geom_geometry, err := geometry.OrbToGeom(ctx, f.Geometry)

	if err != nil {
		return nil, fmt.Errorf("Failed to convert orb geometry to geom geometry, %w", err)
	}

	geom_bounds, err := geometry.OrbToGeom(ctx, bounds.ToPolygon())

	if err != nil {
		return nil, fmt.Errorf("Failed to convert orb bounds to geom geometry, %w", err)
	}

	geom_extent, err := geom.NewExtentFromGeometry(geom_bounds)

	if err != nil {
		return nil, fmt.Errorf("Failed to create extent from bounds, %w", err)
	}

	geom_cropped, err := clip.Geometry(ctx, geom_geometry, geom_extent)

	if err != nil {
		return nil, fmt.Errorf("Failed to crop geometry, %w", err)
	}

	orb_cropped, err := geometry.GeomToOrb(ctx, geom_cropped)

	if err != nil {
		return nil, fmt.Errorf("Failed to convert geom geometry to orb geometry, %w", err)
	}

	/*

		clipped_geom := clip.Geometry(bounds, geom)

		if clipped_geom == nil {
			log.Println("CROP", bounds.Min.Lon(), bounds.Min.Lat(), bounds.Max.Lon(), bounds.Max.Lat())
			log.Println("FROM", geom.Bound())
			return nil, fmt.Errorf("Failed to derive clipped geometry")
		}

		f.Geometry = clipped_geom
	*/

	f.Geometry = orb_cropped
	return f.MarshalJSON()
}
