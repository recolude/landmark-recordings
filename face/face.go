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

var landmarkIrises = []Vector2Int{
	NewVector2Int(475, 476),
	NewVector2Int(477, 474),
	NewVector2Int(469, 470),
	NewVector2Int(472, 469),
	NewVector2Int(471, 472),
	NewVector2Int(474, 475),
	NewVector2Int(476, 477),
	NewVector2Int(470, 471),
}

var landmarkContours = []Vector2Int{
	NewVector2Int(270, 409),
	NewVector2Int(176, 149),
	NewVector2Int(37, 0),
	NewVector2Int(84, 17),
	NewVector2Int(318, 324),
	NewVector2Int(293, 334),
	NewVector2Int(386, 385),
	NewVector2Int(7, 163),
	NewVector2Int(33, 246),
	NewVector2Int(17, 314),
	NewVector2Int(374, 380),
	NewVector2Int(251, 389),
	NewVector2Int(390, 373),
	NewVector2Int(267, 269),
	NewVector2Int(295, 285),
	NewVector2Int(389, 356),
	NewVector2Int(173, 133),
	NewVector2Int(33, 7),
	NewVector2Int(377, 152),
	NewVector2Int(158, 157),
	NewVector2Int(405, 321),
	NewVector2Int(54, 103),
	NewVector2Int(263, 466),
	NewVector2Int(324, 308),
	NewVector2Int(67, 109),
	NewVector2Int(409, 291),
	NewVector2Int(157, 173),
	NewVector2Int(454, 323),
	NewVector2Int(388, 387),
	NewVector2Int(78, 191),
	NewVector2Int(148, 176),
	NewVector2Int(311, 310),
	NewVector2Int(39, 37),
	NewVector2Int(249, 390),
	NewVector2Int(144, 145),
	NewVector2Int(402, 318),
	NewVector2Int(80, 81),
	NewVector2Int(310, 415),
	NewVector2Int(153, 154),
	NewVector2Int(384, 398),
	NewVector2Int(397, 365),
	NewVector2Int(234, 127),
	NewVector2Int(103, 67),
	NewVector2Int(282, 295),
	NewVector2Int(338, 297),
	NewVector2Int(378, 400),
	NewVector2Int(127, 162),
	NewVector2Int(321, 375),
	NewVector2Int(375, 291),
	NewVector2Int(317, 402),
	NewVector2Int(81, 82),
	NewVector2Int(154, 155),
	NewVector2Int(91, 181),
	NewVector2Int(334, 296),
	NewVector2Int(297, 332),
	NewVector2Int(269, 270),
	NewVector2Int(150, 136),
	NewVector2Int(109, 10),
	NewVector2Int(356, 454),
	NewVector2Int(58, 132),
	NewVector2Int(312, 311),
	NewVector2Int(152, 148),
	NewVector2Int(415, 308),
	NewVector2Int(161, 160),
	NewVector2Int(296, 336),
	NewVector2Int(65, 55),
	NewVector2Int(61, 146),
	NewVector2Int(78, 95),
	NewVector2Int(380, 381),
	NewVector2Int(398, 362),
	NewVector2Int(361, 288),
	NewVector2Int(246, 161),
	NewVector2Int(162, 21),
	NewVector2Int(0, 267),
	NewVector2Int(82, 13),
	NewVector2Int(132, 93),
	NewVector2Int(314, 405),
	NewVector2Int(10, 338),
	NewVector2Int(178, 87),
	NewVector2Int(387, 386),
	NewVector2Int(381, 382),
	NewVector2Int(70, 63),
	NewVector2Int(61, 185),
	NewVector2Int(14, 317),
	NewVector2Int(105, 66),
	NewVector2Int(300, 293),
	NewVector2Int(382, 362),
	NewVector2Int(88, 178),
	NewVector2Int(185, 40),
	NewVector2Int(46, 53),
	NewVector2Int(284, 251),
	NewVector2Int(400, 377),
	NewVector2Int(136, 172),
	NewVector2Int(323, 361),
	NewVector2Int(13, 312),
	NewVector2Int(21, 54),
	NewVector2Int(172, 58),
	NewVector2Int(373, 374),
	NewVector2Int(163, 144),
	NewVector2Int(276, 283),
	NewVector2Int(53, 52),
	NewVector2Int(365, 379),
	NewVector2Int(379, 378),
	NewVector2Int(146, 91),
	NewVector2Int(263, 249),
	NewVector2Int(283, 282),
	NewVector2Int(87, 14),
	NewVector2Int(145, 153),
	NewVector2Int(155, 133),
	NewVector2Int(93, 234),
	NewVector2Int(66, 107),
	NewVector2Int(95, 88),
	NewVector2Int(159, 158),
	NewVector2Int(52, 65),
	NewVector2Int(332, 284),
	NewVector2Int(40, 39),
	NewVector2Int(191, 80),
	NewVector2Int(63, 105),
	NewVector2Int(181, 84),
	NewVector2Int(466, 388),
	NewVector2Int(149, 150),
	NewVector2Int(288, 397),
	NewVector2Int(160, 159),
	NewVector2Int(385, 384),
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

		middleConnections := vertConnections[connections[middleIndex].Index]

		for _, endVertex := range middleConnections {

			if endVertex.Done {
				continue
			}

			// Check if end vertex is contaied in original list.
			for end := middleIndex + 1; end < len(connections); end++ {
				// Their not connected
				if endVertex != connections[end] {
					continue
				}

				tris = append(tris, []int{vertToProcess, connections[middleIndex].Index, endVertex.Index})
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

func clockwise(tri []int, firstFrame []position.Capture) bool {
	a := firstFrame[tri[0]].Position()
	b := firstFrame[tri[1]].Position()
	c := firstFrame[tri[2]].Position()

	// N = (B - A) x (C - A)
	n := b.Sub(a).Cross(c.Sub(a))

	// Winding when viewed from V
	// w = N . (A - V)
	return n.Dot(a.Sub(vector.Vector3Forward())) > 0
}

// tesselate is a pretty poor function I'm writing drunk just to get this done.
// There are probably most definantly better ways to do this.
func tesselate(firstFrame []position.Capture) [][]int {
	numVerts := 467 + 1

	vertLUT := make(map[int]*Vertex)
	for i := 0; i < numVerts; i++ {
		vertLUT[i] = NewVertex(i)
	}

	vertConnections := make(map[int][]*Vertex)
	for _, line := range lineSegments {
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

	for _, tri := range tris {
		if clockwise(tri, firstFrame) {
			tri[0], tri[2] = tri[2], tri[0]
		}
	}

	return tris
}

type RunningData struct {
	captures [][]position.Capture
}

func (rd *RunningData) toRecording() format.Recording {
	childrenRecordings := make([]format.Recording, len(rd.captures))

	childStyling := metadata.EmptyBlock()
	// childStyling.Mapping()["recolude-scale"] = metadata.NewStringProperty("0.015, 0.015, 0.015")
	// childStyling.Mapping()["recolude-color"] = metadata.NewStringProperty("#00FFFF")
	childStyling.Mapping()["recolude-geom"] = metadata.NewStringProperty("none")

	for i, col := range rd.captures {
		childrenRecordings[i] = format.NewRecording(
			strconv.Itoa(i),
			strconv.Itoa(i),
			[]format.CaptureCollection{
				position.NewCollection("Position", col),
			},
			nil,
			childStyling,
			nil,
			nil,
		)
	}

	tris := tesselate(rd.FirstFrame())
	allTris := make([]string, len(tris)*3)
	for i, tri := range tris {
		offsetI := i * 3
		allTris[offsetI] = strconv.Itoa(tri[0])
		allTris[offsetI+1] = strconv.Itoa(tri[1])
		allTris[offsetI+2] = strconv.Itoa(tri[2])
	}
	metadataMeshes := []metadata.Block{
		metadata.NewBlock(map[string]metadata.Property{
			"type": metadata.NewStringProperty("subject-as-vertices"),
			"tris": metadata.NewStringArrayProperty(allTris),
		}),
	}

	metadataLines := make([]metadata.Block, 0, len(landmarkContours)+len(landmarkIrises))
	for _, link := range landmarkContours {
		mapping := map[string]metadata.Property{
			"starting-object-id": metadata.NewStringProperty(strconv.Itoa(int(link.X))),
			"ending-object-id":   metadata.NewStringProperty(strconv.Itoa(int(link.Y))),
			"color":              metadata.NewStringProperty("#00FF00"),
			"width":              metadata.NewFloat32Property(0.01),
		}
		metadataLines = append(metadataLines, metadata.NewBlock(mapping))
	}
	for _, link := range landmarkIrises {
		mapping := map[string]metadata.Property{
			"starting-object-id": metadata.NewStringProperty(strconv.Itoa(int(link.X))),
			"ending-object-id":   metadata.NewStringProperty(strconv.Itoa(int(link.Y))),
			"color":              metadata.NewStringProperty("#FF0000"),
			"width":              metadata.NewFloat32Property(0.0025),
		}
		metadataLines = append(metadataLines, metadata.NewBlock(mapping))
	}
	recordingMetadata := metadata.EmptyBlock()
	recordingMetadata.Mapping()["recolude-lines"] = metadata.NewMetadataArrayProperty(metadataLines)
	recordingMetadata.Mapping()["recolude-meshes"] = metadata.NewMetadataArrayProperty(metadataMeshes)
	recordingMetadata.Mapping()["recolude-sun-position"] = metadata.NewVector3Property(0, 200, -100)
	recordingMetadata.Mapping()["recolude-grid"] = metadata.NewStringProperty("false")

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

	// we initialize our Users array
	var frames [][]LandMark

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
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
	recordingWriter.Write(rd.toRecording())
}
