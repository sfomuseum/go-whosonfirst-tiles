// list will emit tile coverage information as CSV-encoded rows for one or more Who's On First records.
package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/sfomuseum/go-whosonfirst-tiles/coverage"
	"github.com/whosonfirst/go-whosonfirst-iterate/iterator"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
)

func main() {

	iter_uri := flag.String("iterator-uri", "repo://", "A valid whosonfirst/go-whosonfirst-iterate/emitter URI.")
	flag.Parse()

	uris := flag.Args()
	ctx := context.Background()

	writers := []io.Writer{
		os.Stdout,
	}

	mw := io.MultiWriter(writers...)
	csv_wr := csv.NewWriter(mw)
	mu := new(sync.RWMutex)

	opts, err := coverage.DefaultCoverageOptions()

	if err != nil {
		log.Fatalf("Failed to create new optsion, %v", err)
	}

	tile_cb := func(ctx context.Context, rsp *coverage.Coverage) error {

		// log.Printf("List tiles for %d at Z%d : %d\n", rsp.Id, rsp.Zoom, len(rsp.Tiles))

		mu.Lock()
		defer mu.Unlock()

		for _, t := range rsp.Tiles {

			out := []string{
				strconv.FormatInt(rsp.Id, 10),
				strconv.Itoa(int(t.Z)),
				strconv.Itoa(int(t.X)),
				strconv.Itoa(int(t.Y)),
			}

			csv_wr.Write(out)
		}

		csv_wr.Flush()
		return nil
	}

	iter_cb := func(ctx context.Context, fh io.ReadSeeker, args ...interface{}) error {

		body, err := io.ReadAll(fh)

		if err != nil {
			return fmt.Errorf("Failed to read record, %v", err)
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

	csv_wr.Flush()
}
