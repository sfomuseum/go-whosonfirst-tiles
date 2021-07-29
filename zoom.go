package tiles

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ZoomLevelsFromString will return a list of uint zoom levels from a string formatted as '1,2,3...' or '1-10'.
func ZoomLevelsFromString(zoom_str string) ([]uint, error) {

	var zooms []uint

	re_zoom, err := regexp.Compile(`^\d+\-\d+$`)

	if err != nil {
		return nil, fmt.Errorf("Failed to compile zoom range regular expression, %w", err)
	}

	if re_zoom.MatchString(zoom_str) {

		zoom_range := strings.Split(zoom_str, "-")

		min_zoom, err := strconv.ParseUint(zoom_range[0], 10, 32)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse min zoom (%s), %w\n", zoom_range[0], err)
		}

		max_zoom, err := strconv.ParseUint(zoom_range[1], 10, 32)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse max zoom (%s), %w\n", zoom_range[1], err)
		}

		if min_zoom > max_zoom {
			return nil, fmt.Errorf("Invalid zoom range")
		}

		zooms = make([]uint, 0)

		for z := min_zoom; z <= max_zoom; z++ {
			zooms = append(zooms, uint(z))
		}

	} else {

		zoomsStrSplit := strings.Split(zoom_str, ",")

		zooms = make([]uint, len(zoomsStrSplit))

		for i, zoomStr := range zoomsStrSplit {

			z, err := strconv.ParseUint(zoomStr, 10, 32)

			if err != nil {
				return nil, fmt.Errorf("Zoom list could not be parsed: %w", err)
			}

			zooms[i] = uint(z)
		}
	}

	return zooms, nil
}
