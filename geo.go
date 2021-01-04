package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang/geo/s2"
	"log"
	"os"
	"strconv"
)
import "github.com/paulmach/go.geojson"

const (
	//EarthRadius the radius of earth in kilometers
	EarthRadius = 6371.01
	maxCells    = 100
)

// Point struct contains the lat/lng of a point
type Point struct {
	Lat float64
	Lng float64
}

var shapeIndex *s2.ShapeIndex
var query *s2.ContainsPointQuery

// DecodeGeoJSON decodes a feature collection
func DecodeGeoJSON(json []byte) ([]*geojson.Feature, error) {
	f, err := geojson.UnmarshalFeatureCollection(json)
	if err != nil {
		return nil, err
	}
	return f.Features, nil
}

// PointsToPolygon converts points to s2 polygon
func PointsToPolygon(points [][]float64) *s2.Polygon {
	var pts []s2.Point
	index := len(points)
	if points[0][0] == points[index-1][0] && points[0][1] == points[index-1][1] {
		index--
	}

	for i := 0; i < index; i++ {
		pts = append(pts, s2.PointFromLatLng(s2.LatLngFromDegrees(points[i][1], points[i][0])))
	}

	loop := s2.LoopFromPoints(pts)

	loop.Normalize()
	err := loop.Validate()
	if err != nil {
		fmt.Printf("loop error %v \n", err)
	}

	return s2.PolygonFromLoops([]*s2.Loop{loop})
}

// CoverPolygon converts s2 polygon to cell union and returns the respective cells
func CoverPolygon(p *s2.Polygon, maxLevel, minLevel int) (s2.CellUnion, []string, [][][]float64) {
	var tokens []string
	var s2cells [][][]float64

	rc := &s2.RegionCoverer{MaxLevel: maxLevel, MinLevel: minLevel, MaxCells: maxCells}
	r := s2.Region(p)
	covering := rc.Covering(r)

	for _, c := range covering {
		cell := s2.CellFromCellID(s2.CellIDFromToken(c.ToToken()))

		s2cells = append(s2cells, EdgesOfCell(cell))

		tokens = append(tokens, c.ToToken())
	}
	return covering, tokens, s2cells
}

// CoverPoint converts a point to cell based on given level
func CoverPoint(p Point, maxLevel int) (s2.Cell, string, [][][]float64) {
	var s2cells [][][]float64

	cid := s2.CellFromLatLng(s2.LatLngFromDegrees(p.Lat, p.Lng)).ID().Parent(maxLevel)
	cell := s2.CellFromCellID(cid)
	token := cid.ToToken()

	s2cells = append(s2cells, EdgesOfCell(cell))

	return cell, token, s2cells
}

// EdgesOfCell gets the edges of the cell
func EdgesOfCell(c s2.Cell) [][]float64 {
	var edges [][]float64
	for i := 0; i < 4; i++ {
		latLng := s2.LatLngFromPoint(c.Vertex(i))
		edges = append(edges, []float64{latLng.Lat.Degrees(), latLng.Lng.Degrees()})
	}
	return edges
}

func IntersectedShapes(point s2.Point) []s2.Shape {
	if query == nil {
		query = s2.NewContainsPointQuery(shapeIndex, s2.VertexModelClosed)
	}

	return query.ContainingShapes(point)
}

func LoadShapeIndex(polygons []*s2.Polygon) {
	if shapeIndex == nil {
		shapeIndex = s2.NewShapeIndex()
	}

	var count int
	for _, polygon := range polygons {
		shapeIndex.Add(polygon)
		if err := polygon.Validate(); err != nil {
			fmt.Printf("error in polygon format: %v \n", err)
			count++
		}
	}
	fmt.Printf("numbers of unprocesed polygons: %d \n", count)
	shapeIndex.Build()
}

//LoadPolygonsFromFile parsing json file as a stream of data => less memory footprint, as we read record by record
func LoadPolygonsFromFile(filename string) (error, []*geojson.Feature) {
	var count int64 = 0
	file, err := os.Open(filename)
	if err != nil {
		return err, nil
	}

	defer file.Close()

	dec := json.NewDecoder(file)

	// read open bracket
	t, err := dec.Token()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%T: %v\n", t, t)

	out := make([]*geojson.Feature, 0)
	// while the array contains values
	for dec.More() {
		m := geojson.Feature{}
		err := dec.Decode(&m)
		out = append(out, &m)
		if err != nil {
			log.Fatal(err)
		}
		count++
	}

	// read closing bracket
	t, err = dec.Token()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%T: %v\n", t, t)
	fmt.Println("--------")
	fmt.Println("count : = " + strconv.FormatInt(count, 10))
	fmt.Println("--------")

	return nil, out
}
