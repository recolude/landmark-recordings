package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"

	"github.com/EliCDavis/vector"
	"github.com/recolude/rap/format"
	"github.com/recolude/rap/format/collection/position"
	"github.com/recolude/rap/format/encoding"
	positionEncoder "github.com/recolude/rap/format/encoding/position"
	rapio "github.com/recolude/rap/format/io"
	"github.com/recolude/rap/format/metadata"
)

type Vector2Int struct {
	X int
	Y int
}

type AABB struct {
	min    vector.Vector3
	max    vector.Vector3
	center *vector.Vector3
}

func NewAABB() *AABB {
	return &AABB{
		min: vector.NewVector3(math.Inf(1), math.Inf(1), math.Inf(1)),
		max: vector.NewVector3(math.Inf(-1), math.Inf(-1), math.Inf(-1)),
	}
}

func (aabb *AABB) Center() vector.Vector3 {
	if aabb.center == nil {
		center := aabb.max.Add(aabb.min).DivByConstant(2)
		aabb.center = &center
	}
	return *aabb.center
}

func (aabb *AABB) CenterPos(x, y, z float64) vector.Vector3 {
	if aabb.center == nil {
		center := aabb.max.Add(aabb.min).DivByConstant(2)
		aabb.center = &center
	}
	return vector.NewVector3(x, y, z).Sub(*aabb.center)
}

func (aabb *AABB) Encompass(x, y, z float64) {
	aabb.center = nil
	aabb.min = vector.NewVector3(
		math.Min(aabb.min.X(), x),
		math.Min(aabb.min.Y(), y),
		math.Min(aabb.min.Z(), z),
	)

	aabb.max = vector.NewVector3(
		math.Max(aabb.max.X(), x),
		math.Max(aabb.max.Y(), y),
		math.Max(aabb.max.Z(), z),
	)
}

type Vertex struct {
	Index int
	Done  bool
}

func (v *Vertex) MarkDone() {
	v.Done = true
}

func NewVertex(index int) *Vertex {
	return &Vertex{Index: index, Done: false}
}

func NewVector2Int(x, y int) Vector2Int {
	return Vector2Int{X: x, Y: y}
}

func process(vertToProcess int, vertConnections map[int][]*Vertex, vertLUT map[int]*Vertex) [][]int {
	connections := vertConnections[vertToProcess]

	tris := make([][]int, 0)

	vertLUT[vertToProcess].MarkDone()

	for middleIndex := 0; middleIndex < len(connections)-1; middleIndex++ {

		// Skip vertices fully explored
		if vertLUT[connections[middleIndex].Index].Done {
			continue
		}

		connectionsToMiddleVertex := vertConnections[connections[middleIndex].Index]

		for _, potentialEndVertex := range connectionsToMiddleVertex {

			if potentialEndVertex.Done {
				continue
			}

			// Check if end vertex is contaied in original list.
			for end := middleIndex + 1; end < len(connections); end++ {
				// Their not connected
				if potentialEndVertex != connections[end] {
					continue
				}

				tris = append(tris, []int{vertToProcess, connections[middleIndex].Index, potentialEndVertex.Index})
				break
			}
		}
	}

	for _, conn := range connections {
		if conn.Done {
			continue
		}
		tris = append(tris, process(conn.Index, vertConnections, vertLUT)...)
	}

	return tris
}

func triNormal(tri []int, firstFrame []position.Capture) vector.Vector3 {
	a := firstFrame[tri[0]].Position()
	b := firstFrame[tri[1]].Position()
	c := firstFrame[tri[2]].Position()
	return b.Sub(a).Cross(c.Sub(a))
}

func clockwise(aabb *AABB, tri []int, firstFrame []position.Capture) bool {
	a := firstFrame[tri[0]].Position()
	b := firstFrame[tri[1]].Position()
	c := firstFrame[tri[2]].Position()

	// N = (B - A) x (C - A)
	n := b.Sub(a).Cross(c.Sub(a))

	// Winding when viewed from V
	// w = N . (A - V)
	return n.Dot(a.Sub(vector.Vector3Forward())) < 0
}

func rewindFaces(tris [][]int, firstFrame []position.Capture) {
	trisProcessed := make(map[int]bool)
	vertsToTris := make([][]int, 467+1)
	for triIndex, tri := range tris {
		vertsToTris[tri[0]] = append(vertsToTris[tri[0]], triIndex)
		vertsToTris[tri[1]] = append(vertsToTris[tri[1]], triIndex)
		vertsToTris[tri[2]] = append(vertsToTris[tri[2]], triIndex)
		trisProcessed[triIndex] = false
	}

	triToNeighbors := make([]map[int]bool, len(tris))
	for triIndex, tri := range tris {
		triToNeighbors[triIndex] = make(map[int]bool)
		for _, vertsTri := range vertsToTris[tri[0]] {
			triToNeighbors[triIndex][vertsTri] = false
		}
		for _, vertsTri := range vertsToTris[tri[1]] {
			triToNeighbors[triIndex][vertsTri] = false
		}
		for _, vertsTri := range vertsToTris[tri[2]] {
			triToNeighbors[triIndex][vertsTri] = false
		}
	}

	triNormals := make([]vector.Vector3, len(tris))
	for triIndex, tri := range tris {
		triNormals[triIndex] = triNormal(tri, firstFrame)
	}

	queue := make([]int, 0)

	// Push to the queue
	queue = append(queue, 0)

	var triIndex int
	for len(queue) > 0 {
		triIndex, queue = queue[0], queue[1:]

		// Skip if we've already been processed in the past
		if trisProcessed[triIndex] {
			continue
		}
		trisProcessed[triIndex] = true

		for neighborIndex := range triToNeighbors[triIndex] {
			if trisProcessed[neighborIndex] {
				continue
			}
			queue = append(queue, neighborIndex)

			otherSide := triNormal([]int{tris[neighborIndex][2], tris[neighborIndex][1], tris[neighborIndex][0]}, firstFrame)
			if triNormals[triIndex].Dot(triNormals[neighborIndex]) < triNormals[triIndex].Dot(otherSide) {
				tris[neighborIndex][0], tris[neighborIndex][2] = tris[neighborIndex][2], tris[neighborIndex][0]
				triNormals[neighborIndex] = otherSide
			}
		}
	}
}

// tesselate is a pretty poor function I'm writing drunk just to get this done.
// There are probably most definantly better ways to do this.
func tesselate(aabb *AABB, firstFrame []position.Capture) [][]int {
	numVerts := 467 + 1

	vertLUT := make(map[int]*Vertex)
	for i := 0; i < numVerts; i++ {
		vertLUT[i] = NewVertex(i)
	}

	vertConnections := make(map[int][]*Vertex)
	processed := make(map[string]bool)
	for _, line := range lineSegments {

		// Avoid processesing duplicate line segments
		id := fmt.Sprintf("%d-%d", line.X, line.Y)
		if line.Y < line.X {
			id = fmt.Sprintf("%d-%d", line.Y, line.X)
		}

		if _, ok := processed[id]; ok {
			continue
		}
		processed[id] = true

		if val, ok := vertConnections[line.X]; ok {
			vertConnections[line.X] = append(val, vertLUT[line.Y])
		} else {
			vertConnections[line.X] = []*Vertex{vertLUT[line.Y]}
		}

		if val, ok := vertConnections[line.Y]; ok {
			vertConnections[line.Y] = append(val, vertLUT[line.X])
		} else {
			vertConnections[line.Y] = []*Vertex{vertLUT[line.X]}
		}
	}

	tris := make([][]int, 0)
	tris = append(tris, process(0, vertConnections, vertLUT)...)

	// Get the tris facing the generally correct direction
	for _, tri := range tris {
		if !clockwise(aabb, tri, firstFrame) {
			tri[0], tri[2] = tri[2], tri[0]
		}
	}

	rewindFaces(tris, firstFrame)

	return tris
}

type RunningData struct {
	captures [][]position.Capture
}

func triIndicesForFace(tris [][]int, faceIndex int) []string {
	offset := 478 * faceIndex
	allTris := make([]string, len(tris)*3)
	for i, tri := range tris {
		offsetI := (i * 3)
		allTris[offsetI] = strconv.Itoa(tri[0] + offset)
		allTris[offsetI+1] = strconv.Itoa(tri[1] + offset)
		allTris[offsetI+2] = strconv.Itoa(tri[2] + offset)
	}
	return allTris
}

func (rd *RunningData) toRecording(aabb *AABB) format.Recording {
	childrenRecordings := make([]format.Recording, len(rd.captures))

	childStyling := metadata.EmptyBlock()
	// childStyling.Mapping()["recolude-scale"] = metadata.NewStringProperty("0.015, 0.015, 0.015")
	// childStyling.Mapping()["recolude-color"] = metadata.NewStringProperty("#00FFFF")
	childStyling.Mapping()["recolude-geom"] = metadata.NewStringProperty("none")

	eyeStyling := metadata.EmptyBlock()
	eyeStyling.Mapping()["recolude-scale"] = metadata.NewStringProperty("0.015, 0.015, 0.015")
	eyeStyling.Mapping()["recolude-color"] = metadata.NewStringProperty("#00FFFF")
	eyeStyling.Mapping()["recolude-geom"] = metadata.NewStringProperty("sphere")
	eyeStyling.Mapping()["body-part"] = metadata.NewStringProperty("pupil")

	for i, col := range rd.captures {
		block := childStyling
		if i%478 == 473 || i%478 == 468 {
			block = eyeStyling
		}

		childrenRecordings[i] = format.NewRecording(
			strconv.Itoa(i),
			strconv.Itoa(i),
			[]format.CaptureCollection{
				position.NewCollection("Position", col),
			},
			nil,
			block,
			nil,
			nil,
		)
	}

	tris := tesselate(aabb, rd.FirstFrame())
	metadataMeshes := make([]metadata.Block, len(rd.captures)/478)
	for faceIndex := 0; faceIndex < len(rd.captures)/478; faceIndex++ {
		metadataMeshes[faceIndex] = metadata.NewBlock(map[string]metadata.Property{
			"type": metadata.NewStringProperty("subject-as-vertices"),
			"tris": metadata.NewStringArrayProperty(triIndicesForFace(tris, faceIndex)),
		})
	}

	metadataLines := make([]metadata.Block, 0, len(landmarkContours)+len(landmarkIrises))

	// Uncomment if you want all the lines around the face
	// for _, link := range landmarkContours {
	// 	mapping := map[string]metadata.Property{
	// 		"starting-object-id": metadata.NewStringProperty(strconv.Itoa(int(link.X))),
	// 		"ending-object-id":   metadata.NewStringProperty(strconv.Itoa(int(link.Y))),
	// 		"color":              metadata.NewStringProperty("#00FF00"),
	// 		"width":              metadata.NewFloat32Property(0.01),
	// 	}
	// 	metadataLines = append(metadataLines, metadata.NewBlock(mapping))
	// }

	for faceIndex := 0; faceIndex < len(rd.captures)/478; faceIndex++ {
		for _, link := range landmarkIrises {
			mapping := map[string]metadata.Property{
				"starting-object-id": metadata.NewStringProperty(strconv.Itoa(int(link.X) + (faceIndex * 478))),
				"ending-object-id":   metadata.NewStringProperty(strconv.Itoa(int(link.Y) + (faceIndex * 478))),
				"color":              metadata.NewStringProperty("#00FFFF"),
				"width":              metadata.NewFloat32Property(0.0025),
			}
			metadataLines = append(metadataLines, metadata.NewBlock(mapping))
		}
	}

	recordingMetadata := metadata.EmptyBlock()
	recordingMetadata.Mapping()["recolude-lines"] = metadata.NewMetadataArrayProperty(metadataLines)
	recordingMetadata.Mapping()["recolude-meshes"] = metadata.NewMetadataArrayProperty(metadataMeshes)
	recordingMetadata.Mapping()["recolude-sun-position"] = metadata.NewVector3Property(0, 200, -100)
	recordingMetadata.Mapping()["recolude-grid"] = metadata.NewStringProperty("false")
	recordingMetadata.Mapping()["recolude-skybox"] = metadata.NewStringProperty("webplayer-assets/examples/landmarks/nightskycolor.png")

	return format.NewRecording(
		"face",
		"Face Capture Demo",
		[]format.CaptureCollection{},
		childrenRecordings,
		recordingMetadata,
		nil,
		nil,
	)
}

func (rd *RunningData) process(curTime float64, frame []LandMark, aabb *AABB) {
	for i, landmark := range frame {
		if len(rd.captures) < i+1 {
			rd.captures = append(rd.captures, make([]position.Capture, 0))
		}
		centered := aabb.CenterPos(landmark.X, landmark.Y, landmark.Z)
		rd.captures[i] = append(rd.captures[i], position.NewCapture(curTime, (centered.X()*2), (-centered.Y()*2), centered.Z()*2))
	}
}

func (rd *RunningData) FirstFrame() []position.Capture {
	firstFrame := make([]position.Capture, len(rd.captures))
	for i, collection := range rd.captures {
		firstFrame[i] = collection[0]
	}
	return firstFrame
}

type LandMark struct {
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	Z  float64 `json:"z"`
	ID int     `json:"id"`
}

func (lm LandMark) String() string {
	return fmt.Sprintf("%d: %f, %f, %f", lm.ID, lm.X, lm.Y, lm.Z)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	jsonFile, err := os.Open("face.json")
	check(err)
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	check(err)

	var frames [][]LandMark
	check(json.Unmarshal(byteValue, &frames))

	// Calc AABB to shift to center
	aabb := NewAABB()
	for _, frame := range frames {
		for _, mark := range frame {
			aabb.Encompass(mark.X, mark.Y, mark.Z)
		}
	}

	rd := &RunningData{
		captures: make([][]position.Capture, 0),
	}
	curTime := 0.0

	for _, frame := range frames {
		rd.process(curTime, frame, aabb)
		curTime += 1.0 / 30.0
	}

	f, _ := os.Create("face tracking.rap")
	recordingWriter := rapio.NewWriter(
		[]encoding.Encoder{
			positionEncoder.NewEncoder(positionEncoder.Oct24),
		},
		true,
		f,
		rapio.BST16,
	)
	recordingWriter.Write(rd.toRecording(aabb))
}
