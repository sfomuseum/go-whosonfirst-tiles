package render

import (
	"context"
	"fmt"
	"github.com/go-spatial/geom"
	"github.com/paulmach/orb/geojson"
	"github.com/whosonfirst/go-geojson-svg"
	"io"
)

type SVGOptions struct {
	TileSize   float64      `json:"tile_size"`
	TileExtent *geom.Extent `json:"tile_extent"`
	Writer     io.Writer
}

func RenderSVGWithFeatures(ctx context.Context, opts *SVGOptions, features ...*geojson.Feature) error {

	s := svg.New()
	s.Mercator = true

	for idx, f := range features {

		enc_f, err := f.MarshalJSON()

		if err != nil {
			return fmt.Errorf("Failed to unmarshal feature (at index %d) to render, %w", idx, err)
		}

		err = s.AddFeature(string(enc_f))

		if err != nil {
			return fmt.Errorf("Failed to add feature (at index %d) to render, %w", idx, err)
		}
	}

	props := make([]string, 0)

	tile_size := opts.TileSize

	if opts.TileExtent != nil {

		s.Extent = &svg.Extent{
			MinX: opts.TileExtent.MinX(),
			MinY: opts.TileExtent.MinY(),
			MaxX: opts.TileExtent.MaxX(),
			MaxY: opts.TileExtent.MaxY(),
		}
	}

	rsp := s.Draw(tile_size, tile_size,
		svg.WithAttribute("xmlns", "http://www.w3.org/2000/svg"),
		svg.WithAttribute("viewBox", fmt.Sprintf("0 0 %d %d", int(tile_size), int(tile_size))),
		svg.UseProperties(props),
	)

	_, err := opts.Writer.Write([]byte(rsp))
	return err
}
