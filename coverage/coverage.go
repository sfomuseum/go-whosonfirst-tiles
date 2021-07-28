// package coverage provides methods for deriving map tile coverage for Who's On First records.
package coverage

import (
	"context"
	"fmt"
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/slippy"
	"github.com/paulmach/orb/geojson"
	_ "log"
	"sync"
)

// CoverageOptions defines common options for the CoverageWithFeature and CoverageWithFeatureAndChannels methods
type CoverageOptions struct {
	// A valid go-spatial/geom/slippy.Grid used to generate coverage information.
	Grid slippy.Grid
	// A list of zoom levels to determine coverage for.
	ZoomLevels []uint
}

// Coverage is a struct containing information returned by the CoverageWithFeatureAndChannels.
type Coverage struct {
	// The Who's On First ID of the record being processed.
	Id int64
	// The zoom level being processed.
	Zoom uint
	// The list of tiles that cover feature 'Id' at zoom level 'Zoom'.
	Tiles []slippy.Tile
}

// CoverageCallbackFunc is a user-defined callback function invoked by CoverageWithFeatureAndCallback method.
type CoverageCallbackFunc func(context.Context, *Coverage) error

// DefaultCoverageOptions returns a CoverageOptions instance with a 4326 grid and zoom levels ranging from 1 to 20.
func DefaultCoverageOptions() (*CoverageOptions, error) {

	grid, err := slippy.NewGrid(4326) // 3857)

	if err != nil {
		return nil, err
	}

	zoom_levels := make([]uint, 0)

	for i := 1; i < 21; i++ {
		zoom_levels = append(zoom_levels, uint(i))
	}

	opts := &CoverageOptions{
		Grid:       grid,
		ZoomLevels: zoom_levels,
	}

	return opts, nil
}

// CoverageWithFeature returns a map of tiles for a Who's On Feature record. The map is keyed by zoom level and the value of each is a list of tiles that cover the feature.
func CoverageWithFeature(ctx context.Context, opts *CoverageOptions, body []byte) (map[uint][]slippy.Tile, error) {

	rsp_ch := make(chan *Coverage)
	err_ch := make(chan error)
	done_ch := make(chan bool)

	go CoverageWithFeatureAndChannels(ctx, opts, body, rsp_ch, err_ch, done_ch)

	t := make(map[uint][]slippy.Tile)

	for {
		select {
		case <-ctx.Done():
			return t, nil
		case <-done_ch:
			return t, nil
		case err := <-err_ch:
			return nil, err
		case rsp := <-rsp_ch:
			t[rsp.Zoom] = rsp.Tiles
		default:
			// pass
		}
	}
}

// CoverageWithFeatureAndCallback will dispatch coverage information for each zoom level defined in 'opts' to a callback function defined in 'cb'.
func CoverageWithFeatureAndCallback(ctx context.Context, opts *CoverageOptions, body []byte, cb CoverageCallbackFunc) error {

	rsp_ch := make(chan *Coverage)
	err_ch := make(chan error)
	done_ch := make(chan bool)

	go CoverageWithFeatureAndChannels(ctx, opts, body, rsp_ch, err_ch, done_ch)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-done_ch:
			return nil
		case err := <-err_ch:
			return err
		case rsp := <-rsp_ch:

			err := cb(ctx, rsp)

			if err != nil {
				return fmt.Errorf("Callback function failed, %w", err)
			}

		default:
			// pass
		}
	}

}

// CoverageWithFeatureAndChannels returns coverage information for each zoom level defined in 'opts' as it is determined using channels.
func CoverageWithFeatureAndChannels(ctx context.Context, opts *CoverageOptions, body []byte, rsp_ch chan *Coverage, err_ch chan error, done_ch chan bool) {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	defer func() {
		done_ch <- true
	}()

	f, err := geojson.UnmarshalFeature(body)

	if err != nil {
		err_ch <- fmt.Errorf("Failed to unmarshal feature, %v", err)
		return
	}

	props := f.Properties
	id_raw, exists := props["wof:id"]

	if !exists {
		err_ch <- fmt.Errorf("Missing wof:id property")
		return
	}

	id := int64(id_raw.(float64))

	bounds := f.Geometry.Bound()

	extent := &geom.Extent{
		bounds.Min.X(),
		bounds.Min.Y(),
		bounds.Max.X(),
		bounds.Max.Y(),
	}

	wg := new(sync.WaitGroup)

	for _, z := range opts.ZoomLevels {

		wg.Add(1)

		go func(z uint) {

			defer wg.Done()

			t := slippy.FromBounds(opts.Grid, extent, z)

			rsp := &Coverage{
				Id:    id,
				Zoom:  z,
				Tiles: t,
			}

			rsp_ch <- rsp
		}(z)
	}

	wg.Wait()
}
