package main

import (
	"fmt"
	"github.com/golang/geo/s2"
	geojson "github.com/paulmach/go.geojson"
	"log"
	"net/http"
	"os"
)

func main() {
	fmt.Println("Starting geo calculation...")

	err, geoJSON := LoadPolygonsFromFile("static/countries-polygons.json")

	if err != nil {
		fmt.Printf("Error -> %v\n", err)
		os.Exit(-1)
	}

	var innerPoint s2.Point
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
				//outerPoint = s2.PointFromLatLng(latLng)
			}

		case geojson.GeometryMultiPolygon:
			//for _, mp := range feature.Geometry.MultiPolygon {
			//	for _, p := range mp {
			//		polygons = append(polygons, PointsToPolygon(p))
			//	}
			//}
		}
	}

	LoadShapeIndex(polygons)
	shapes := IntersectedShapes(innerPoint)
	polygon := make([][][]float64, 1)
	var shapeCount int
	for _, shape := range shapes {
		for _, loop := range shape.(*s2.Polygon).Loops() {
			for _, point := range loop.Vertices() {
				polygon[shapeCount] = append(polygon[shapeCount],
					[]float64{
						s2.LatLngFromPoint(point).Lng.Degrees(), s2.LatLngFromPoint(point).Lat.Degrees(),
					},
				)
			}
		}
		shapeCount++
	}

	fmt.Printf("Number of intersected polygons: %d\n", shapeCount)

	jsonOutput, _ := geojson.NewPolygonFeature(polygon).MarshalJSON()
	fmt.Printf("OUTPUT intersection: %s", string(jsonOutput))

}

func StartWebServer() {
	fmt.Println("Starting Leaflet web page...")

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	log.Println("Listening on :3000...")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
