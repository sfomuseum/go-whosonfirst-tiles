package render

import (
	"context"
	"fmt"
	"github.com/go-spatial/geom"
	"github.com/paulmach/orb/geojson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-geojson-svg"
	"io"
	"strconv"
)

// SVGOptions defines common configuration options for the RenderSVGWithFeatures method.
type SVGOptions struct {
	// The size of the tile to render
	TileSize float64 `json:"tile_size"`
	// An optional extent to assign the final SVG output.
	TileExtent *geom.Extent `json:"tile_extent"`
	// A valid io.Writer where SVG data will be written to.
	Writer io.Writer
	// A valid SVG stroke value.
	Stroke string `json:"stroke"`
	// A valid SVG stroke-width value.
	StrokeWidth float64 `json:"stroke_width"`
	// A valid SVG stroke-opacity value.
	StrokeOpacity float64 `json:"stroke_opacity"`
	// A valid SVG fill value.
	Fill string `json:"fill"`
	// A valid SVG fill-opacity value.
	FillOpacity float64 `json:"fill_opacity"`
}

// DefaultSVGOptions returns default configuration options for using with the DefaultSVGOptions method.
func DefaultSVGOptions() *SVGOptions {

	opts := &SVGOptions{
		TileSize:      512,
		Stroke:        "#000000",
		StrokeWidth:   1.0,
		StrokeOpacity: 1.0,
		Fill:          "#ffffff",
		FillOpacity:   0.0,
		Writer:        io.Discard,
	}

	return opts
}

// Render SVG data for one or more geojson.Feature instances.
func RenderSVGWithFeatures(ctx context.Context, opts *SVGOptions, features ...*geojson.Feature) error {

	s := svg.New()
	s.Mercator = true

	stroke := opts.Stroke
	stroke_width := opts.StrokeWidth
	stroke_opacity := opts.StrokeOpacity

	fill := opts.Fill
	fill_opacity := opts.FillOpacity

	use_props := map[string]interface{}{
		"stroke":         stroke,
		"fill":           fill,
		"stroke-width":   strconv.FormatFloat(stroke_width, 'f', -1, 64),
		"stroke-opacity": strconv.FormatFloat(stroke_opacity, 'f', -1, 64),
		"fill-opacity":   strconv.FormatFloat(fill_opacity, 'f', -1, 64),
	}

	for idx, f := range features {

		enc_f, err := f.MarshalJSON()

		if err != nil {
			return fmt.Errorf("Failed to unmarshal feature (at index %d) to render, %w", idx, err)
		}

		for k, v := range use_props {
			path := fmt.Sprintf("properties.%s", k)
			enc_f, _ = sjson.SetBytes(enc_f, path, v)
		}

		err = s.AddFeature(string(enc_f))

		if err != nil {
			return fmt.Errorf("Failed to add feature (at index %d) to render, %w", idx, err)
		}
	}

	tile_size := opts.TileSize

	if opts.TileExtent != nil {

		s.Extent = &svg.Extent{
			MinX: opts.TileExtent.MinX(),
			MinY: opts.TileExtent.MinY(),
			MaxX: opts.TileExtent.MaxX(),
			MaxY: opts.TileExtent.MaxY(),
		}
	}

	props := make([]string, 0)

	for k, _ := range use_props {
		props = append(props, k)
	}

	rsp := s.Draw(tile_size, tile_size,
		svg.WithAttribute("xmlns", "http://www.w3.org/2000/svg"),
		svg.WithAttribute("viewBox", fmt.Sprintf("0 0 %d %d", int(tile_size), int(tile_size))),
		svg.UseProperties(props),
	)

	_, err := opts.Writer.Write([]byte(rsp))
	return err
}
