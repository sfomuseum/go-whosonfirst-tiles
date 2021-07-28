package render

import (
	"context"
	"fmt"
	"github.com/go-spatial/geom"
	"github.com/whosonfirst/go-geojson-svg"
	"io"
)

type SVGOptions struct {
	TileSize   float64      `json:"tile_size"`
	TileExtent *geom.Extent `json:"tile_extent"`
	Writer     io.Writer
}

func RenderSVG(ctx context.Context, opts *SVGOptions, body []byte) error {

	s := svg.New()
	s.Mercator = true

	err := s.AddFeature(string(body))

	if err != nil {
		return fmt.Errorf("Failed to add feature to render, %w", err)
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

	_, err = opts.Writer.Write([]byte(rsp))
	return err
}
