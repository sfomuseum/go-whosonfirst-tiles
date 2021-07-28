// render will generate ... tiles for one or more Who's On First records.
package main

import (
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"
)

import (
	"context"
	"flag"
	"fmt"
	"github.com/paulmach/orb/geojson"
	"github.com/sfomuseum/go-whosonfirst-tiles/coverage"
	"github.com/sfomuseum/go-whosonfirst-tiles/crop"
	"github.com/sfomuseum/go-whosonfirst-tiles/render"
	"github.com/whosonfirst/go-whosonfirst-iterate/iterator"
	"gocloud.dev/blob"
	"io"
	"log"
)

func main() {

	bucket_uri := flag.String("bucket-uri", "mem://", "A valid gocloud.dev/blob URI.")
	iter_uri := flag.String("iterator-uri", "repo://", "A valid whosonfirst/go-whosonfirst-iterate/emitter URI.")
	flag.Parse()

	uris := flag.Args()
	ctx := context.Background()

	bucket, err := blob.OpenBucket(ctx, *bucket_uri)

	if err != nil {
		log.Fatalf("Failed to open bucket, %v", err)
	}

	defer bucket.Close()

	opts, err := coverage.DefaultCoverageOptions()

	if err != nil {
		log.Fatalf("Failed to create new optsion, %v", err)
	}

	opts.ZoomLevels = []uint{15}

	iter_cb := func(ctx context.Context, fh io.ReadSeeker, args ...interface{}) error {

		body, err := io.ReadAll(fh)

		if err != nil {
			return fmt.Errorf("Failed to read record, %v", err)
		}

		f, err := geojson.UnmarshalFeature(body)

		if err != nil {
			return fmt.Errorf("Failed to unmarshal feature, %w", err)
		}

		tile_cb := func(ctx context.Context, rsp *coverage.Coverage) error {

			for _, t := range rsp.Tiles {

				fname := fmt.Sprintf("%d/%d/%d.svg", t.Z, t.X, t.Y)

				cropped, err := crop.CropFeatureWithTile(ctx, f, &t, opts.Grid)

				if err != nil {
					return fmt.Errorf("Failed to crop feature, %w", err)
				}

				wr, err := bucket.NewWriter(ctx, fname, nil)

				if err != nil {
					return fmt.Errorf("Failed to create new writer for '%s', %v", fname, err)
				}

				svg_opts := &render.SVGOptions{
					Writer: wr,
				}

				err = render.RenderSVG(ctx, svg_opts, cropped)

				if err != nil {
					return fmt.Errorf("Failed to render '%s', %v", fname, err)
				}

				return wr.Close()
			}

			return nil
		}

		return coverage.CoverageWithFeatureAndCallback(ctx, opts, body, tile_cb)
	}

	iter, err := iterator.NewIterator(ctx, *iter_uri, iter_cb)

	if err != nil {
		log.Fatalf("Failed to create new iterator, %v", err)
	}

	err = iter.IterateURIs(ctx, uris...)

	if err != nil {
		log.Fatalf("Failed to iterator URIs, %v", err)
	}

}
