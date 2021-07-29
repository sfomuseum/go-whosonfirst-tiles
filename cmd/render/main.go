// render will generate tiles for one or more Who's On First records. This tool uses a two-pass approach. The first
// pass collects all the cropped features associated with the map tiles for a given record and stores them as a GeoJSON
// FeatureCollection. The second pass will iterate over those FeatureCollection records (associated with a map tile) and
// generate a corresponding SVG file.
package main

import (
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"
)

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-spatial/geom/slippy"
	"github.com/paulmach/orb/geojson"
	"github.com/sfomuseum/go-whosonfirst-tiles"
	"github.com/sfomuseum/go-whosonfirst-tiles/coverage"
	"github.com/sfomuseum/go-whosonfirst-tiles/crop"
	"github.com/sfomuseum/go-whosonfirst-tiles/render"
	"github.com/whosonfirst/go-whosonfirst-iterate/iterator"
	"gocloud.dev/blob"
	"io"
	"log"
	_ "os"
	"regexp"
	"strconv"
	"sync"
)

func main() {

	data_bucket_uri := flag.String("data-bucket-uri", "mem://", "A valid gocloud.dev/blob URI for writing intermediate data records.")
	tile_bucket_uri := flag.String("tile-bucket-uri", "mem://", "A valid gocloud.dev/blob URI for writing tile data.")

	iter_uri := flag.String("iterator-uri", "repo://", "A valid whosonfirst/go-whosonfirst-iterate/emitter URI.")

	zoom_str := flag.String("zoom-levels", "10-18", "Comma-separated list of zoom levels or a '{MIN_ZOOM}-{MAX_ZOOM}' range string.")

	flag.Parse()

	uris := flag.Args()
	ctx := context.Background()

	data_bucket, err := blob.OpenBucket(ctx, *data_bucket_uri)

	if err != nil {
		log.Fatalf("Failed to open bucket, %v", err)
	}

	defer data_bucket.Close()

	tile_bucket, err := blob.OpenBucket(ctx, *tile_bucket_uri)

	if err != nil {
		log.Fatalf("Failed to open bucket, %v", err)
	}

	defer tile_bucket.Close()

	coverage_opts, err := coverage.DefaultCoverageOptions()

	if err != nil {
		log.Fatalf("Failed to create new optsion, %v", err)
	}

	zoom_levels, err := tiles.ZoomLevelsFromString(*zoom_str)

	if err != nil {
		log.Fatalf("Failed to derive zoom levels, %v", err)
	}

	coverage_opts.ZoomLevels = zoom_levels

	// Step 1: Gather all the tile data to render

	mu := new(sync.RWMutex)

	iter_cb := func(ctx context.Context, fh io.ReadSeeker, args ...interface{}) error {

		body, err := io.ReadAll(fh)

		if err != nil {
			return fmt.Errorf("Failed to read record, %v", err)
		}

		tile_cb := func(ctx context.Context, rsp *coverage.Coverage) error {

			for t, _ := range rsp.Tiles {

				path := fmt.Sprintf("%d/%d/%d.geojson", t.Z, t.X, t.Y)
				// log.Println(path)

				cropped, err := crop.CropFeatureWithTile(ctx, body, t)

				// This seems to be rooted in the orb/clip/clip.go ring()
				// method which keeps returning nil but I don't know why
				// yet...

				if err != nil {
					log.Printf("Failed to crop feature '%s', %v", path, err)
					continue

					// return fmt.Errorf("Failed to crop feature, %w", err)
				}

				cropped_f, err := geojson.UnmarshalFeature(cropped)

				if err != nil {
					return fmt.Errorf("Failed to unmarshal cropped feature for '%s', %w", path, err)
				}

				mu.Lock()
				defer mu.Unlock()

				exists, err := data_bucket.Exists(ctx, path)

				if err != nil {
					return fmt.Errorf("Failed to determine whether '%s' exists, %w", path, err)
				}

				var fc *geojson.FeatureCollection

				if exists {

					fh, err := data_bucket.NewReader(ctx, path, nil)

					if err != nil {
						return fmt.Errorf("Failed to open '%s', %w", path, err)
					}

					defer fh.Close()

					body, err := io.ReadAll(fh)

					if err != nil {
						return fmt.Errorf("Failed to read '%s', %w", path, err)
					}

					doc, err := geojson.UnmarshalFeatureCollection(body)

					if err != nil {
						return fmt.Errorf("Failed to unmarshal '%s', %w", path, err)
					}

					fc = doc
				} else {
					fc = geojson.NewFeatureCollection()
				}

				fc.Append(cropped_f)

				enc_fc, err := fc.MarshalJSON()

				if err != nil {
					return fmt.Errorf("Failed to marshal '%s', %w", path, err)
				}

				wr, err := data_bucket.NewWriter(ctx, path, nil)

				if err != nil {
					return fmt.Errorf("Failed to create new writer for '%s', %v", path, err)
				}

				_, err = wr.Write(enc_fc)

				if err != nil {
					return fmt.Errorf("Failed to write '%s', %w", path, err)
				}

				return wr.Close()
			}

			return nil
		}

		return coverage.CoverageWithFeatureAndCallback(ctx, coverage_opts, body, tile_cb)
	}

	iter, err := iterator.NewIterator(ctx, *iter_uri, iter_cb)

	if err != nil {
		log.Fatalf("Failed to create new iterator, %v", err)
	}

	err = iter.IterateURIs(ctx, uris...)

	if err != nil {
		log.Fatalf("Failed to iterator URIs, %v", err)
	}

	// Step 1: Render the tile data

	re, err := regexp.Compile(`(\d+)\/(\d+)\/(\d+)\.geojson$`)

	if err != nil {
		log.Fatalf("Failed to compile tile regular expression, %v", err)
	}

	var list func(context.Context, *blob.Bucket, string) error

	list = func(ctx context.Context, data_bucket *blob.Bucket, prefix string) error {

		iter := data_bucket.List(&blob.ListOptions{
			Delimiter: "/",
			Prefix:    prefix,
		})

		for {
			obj, err := iter.Next(ctx)

			if err == io.EOF {
				break
			}

			if err != nil {
				return err
			}

			path := obj.Key

			if obj.IsDir {

				err := list(ctx, data_bucket, path)

				if err != nil {
					return err
				}

				continue
			}

			//

			m := re.FindStringSubmatch(path)

			if len(m) == 0 {
				continue
			}

			fh, err := data_bucket.NewReader(ctx, path, nil)

			if err != nil {
				return fmt.Errorf("Failed to open '%s', %v", path, err)
			}

			defer fh.Close()

			body, err := io.ReadAll(fh)

			if err != nil {
				return fmt.Errorf("Failed to read '%s', %v", path, err)
			}

			fc, err := geojson.UnmarshalFeatureCollection(body)

			if err != nil {
				return fmt.Errorf("Failed to unmarshal '%s', %v", path, err)
			}

			str_z := m[1]
			str_x := m[2]
			str_y := m[3]

			z, _ := strconv.Atoi(str_z)
			x, _ := strconv.Atoi(str_x)
			y, _ := strconv.Atoi(str_y)

			t_path := fmt.Sprintf("%d/%d/%d.svg", z, x, y)

			// replace with maptile.Tile?
			t := slippy.NewTile(uint(z), uint(x), uint(y))

			wr, err := tile_bucket.NewWriter(ctx, t_path, nil)

			if err != nil {
				return fmt.Errorf("Failed to create new writer for '%s', %v", t_path, err)
			}

			extent := tiles.Extent4326(t)

			svg_opts := render.DefaultSVGOptions()
			svg_opts.TileExtent = extent
			svg_opts.Writer = wr

			err = render.RenderSVGWithFeatures(ctx, svg_opts, fc.Features...)

			if err != nil {
				return fmt.Errorf("Failed to render '%s', %v", t_path, err)
			}

			err = wr.Close()

			if err != nil {
				return fmt.Errorf("Failed to close '%s', %v", t_path, err)
			}

			log.Println("Wrote", t_path)

			//

			err = data_bucket.Delete(ctx, path)

			if err != nil {
				log.Printf("Failed to delete '%s', %v", path, err)
			}
		}

		return nil
	}

	err = list(ctx, data_bucket, "")

	if err != nil {
		log.Fatalf("Failed to list data bucket, %v", err)
	}

}
