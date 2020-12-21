package main

import (
	"fmt"
	"github.com/golang/geo/s2"
	geojson "github.com/paulmach/go.geojson"
	"os"
)

func main() {
	fmt.Println("Starting geo calculation...")

	gb := []byte(`{
  "type": "FeatureCollection",
  "features": [
    {
      "type": "Feature",
      "properties": {},
      "geometry": {
        "type": "Polygon",
        "coordinates": [
          [
            [
              -56.180992126464844,
              -34.90057413710918
            ],
            [
              -56.13945007324218,
              -34.90057413710918
            ],
            [
              -56.13945007324218,
              -34.872693498558775
            ],
            [
              -56.180992126464844,
              -34.872693498558775
            ],
            [
              -56.180992126464844,
              -34.90057413710918
            ]
          ]
        ]
      }
    },
    {
      "type": "Feature",
      "properties": {
        "inner": true
      },
      "geometry": {
        "type": "Point",
        "coordinates": [
          -56.15867614746093,
          -34.885367688770835
        ]
      }
    },
    {
      "type": "Feature",
      "properties": {
        "inner": false
      },
      "geometry": {
        "type": "Point",
        "coordinates": [
          -56.12297058105469,
          -34.88086153393072
        ]
      }
    }
  ]
}`)

	geoJSON, err := DecodeGeoJSON(gb)
	if err != nil {
		fmt.Printf("Error -> %v\n", err)
		os.Exit(-1)
	}

	var outerPoint, innerPoint s2.Point
	polygons := make([]*s2.Polygon, 0)

	for _, feature := range geoJSON {

		switch feature.Geometry.Type {
		case geojson.GeometryPolygon:
			for _, p := range feature.Geometry.Polygon {
				polygons = append(polygons, PointsToPolygon(p))
			}

		case geojson.GeometryPoint:
			latLng := s2.LatLngFromDegrees(feature.Geometry.Point[1], feature.Geometry.Point[0])
			if feature.Properties["inner"].(bool) {
				innerPoint = s2.PointFromLatLng(latLng)
			} else {
				outerPoint = s2.PointFromLatLng(latLng)
			}
		}
	}

	shapes := IntersectedShapes(innerPoint, polygons)
	polygon := make([][][]float64, 1)
	for _, shape := range shapes {
		for _, loop := range shape.(*s2.Polygon).Loops() {
			for _, point := range loop.Vertices() {
				fmt.Printf("Points: %v\n", s2.LatLngFromPoint(point))
				polygon[0] = append(polygon[0],
					[]float64{
						s2.LatLngFromPoint(point).Lng.Degrees(), s2.LatLngFromPoint(point).Lat.Degrees(),
					},
				)
			}
		}

	}

	json, _ := geojson.NewPolygonFeature(polygon).MarshalJSON()

	fmt.Printf("OUTPUT intersection: %s", string(json))
	shapes2 := IntersectedShapes(outerPoint, polygons)

	for _, shape := range shapes2 {
		fmt.Printf("Outer Point: %v", shape)
	}
}
