module github.com/sfomuseum/go-whosonfirst-tiles

go 1.16

require (
	github.com/go-spatial/geom v0.0.0-20210315165355-0e06498b3362
	github.com/paulmach/orb v0.2.2
	github.com/sfomuseum/go-csvdict v0.0.1
	github.com/whosonfirst/go-whosonfirst-iterate v1.2.0
)

// https://github.com/go-spatial/geom/pull/117
replace github.com/go-spatial/geom v0.0.0-20210315165355-0e06498b3362 => github.com/sfomuseum/geom v0.0.0-20210728180130-14e8d02880d3
