package geometry

import (
	"bytes"
	"context"
	"github.com/go-spatial/geom"
	geom_wkb "github.com/go-spatial/geom/encoding/wkb"
	geom_wkt "github.com/go-spatial/geom/encoding/wkt"
	"github.com/paulmach/orb"
	orb_wkb "github.com/paulmach/orb/encoding/wkb"
	orb_wkt "github.com/paulmach/orb/encoding/wkt"
)

func OrbToGeom(ctx context.Context, g orb.Geometry) (geom.Geometry, error) {
	str_wkt := orb_wkt.MarshalString(g)
	return geom_wkt.DecodeString(str_wkt)
}

func GeomToOrb(ctx context.Context, g geom.Geometry) (orb.Geometry, error) {

	wkb_body, err := geom_wkb.EncodeBytes(g)

	if err != nil {
		return nil, err
	}

	br := bytes.NewReader(wkb_body)

	dec := orb_wkb.NewDecoder(br)
	return dec.Decode()

}
