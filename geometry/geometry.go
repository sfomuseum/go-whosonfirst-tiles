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
	"github.com/paulsmith/gogeos/geos"	
	"log"
)

func OrbToGeom(ctx context.Context, g orb.Geometry) (geom.Geometry, error) {
	str_wkt := orb_wkt.MarshalString(g)
	return geom_wkt.DecodeString(str_wkt)
}

func GeomToOrb(ctx context.Context, g geom.Geometry) (orb.Geometry, error) {

	log.Println("WHAT", g)
	wkb_body, err := geom_wkb.EncodeBytes(g)

	if err != nil {
		return nil, err
	}

	br := bytes.NewReader(wkb_body)

	dec := orb_wkb.NewDecoder(br)
	return dec.Decode()

}

func OrbToGeos(ctx context.Context, orb_geom orb.Geometry) (*geos.Geometry, error) {
	str_wkt := orb_wkt.MarshalString(orb_geom)
	return geos.FromWKT(str_wkt)
}

func GeosToOrb(ctx context.Context, geos_geom *geos.Geometry) (orb.Geometry, error) {

	wkb_body, err := geos_geom.WKB()

	if err != nil {
		return nil, err
	}

	br := bytes.NewReader(wkb_body)

	dec := orb_wkb.NewDecoder(br)
	return dec.Decode()
}
