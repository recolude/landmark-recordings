package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/recolude/rap/format"
	"github.com/recolude/rap/format/collection/position"
	"github.com/recolude/rap/format/encoding"
	positionEncoder "github.com/recolude/rap/format/encoding/position"
	rapio "github.com/recolude/rap/format/io"
	"github.com/recolude/rap/format/metadata"
)

var landmarkNames = []string{
	"NOSE",
	"LEFT EYE_INNER",
	"LEFT EYE",
	"LEFT EYE OUTER",
	"RIGHT EYE INNER",
	"RIGHT EYE",
	"RIGHT EYE OUTER",
	"LEFT EAR",
	"RIGHT EAR",
	"MOUTH LEFT",
	"MOUTH RIGHT",

	"LEFT SHOULDER",
	"RIGHT SHOULDER",
	"LEFT ELBOW",
	"RIGHT ELBOW",
	"LEFT WRIST",
	"RIGHT WRIST",
	"LEFT PINKY",
	"RIGHT PINKY",
	"LEFT INDEX",
	"RIGHT INDEX",
	"LEFT THUMB",
	"RIGHT THUMB",
	"LEFT HIP",
	"RIGHT HIP",
	"LEFT KNEE",
	"RIGHT KNEE",
	"LEFT ANKLE",
	"RIGHT ANKLE",
	"LEFT HEEL",
	"RIGHT HEEL",
	"LEFT FOOT INDEX",
	"RIGHT FOOT INDEX",
}

var landmarkColors = []string{
	"#FF0000",
	"#FFA500",
	"#FFA500",
	"#FFA500",
	"#00FFFF",
	"#00FFFF",
	"#00FFFF",
	"#FFA500",
	"#00FFFF",
	"#FFA500",
	"#00FFFF",

	"#FFA500",
	"#00FFFF",
	"#FFA500",
	"#00FFFF",
	"#FFA500",
	"#00FFFF",
	"#FFA500",
	"#00FFFF",
	"#FFA500",
	"#00FFFF",
	"#FFA500",
	"#00FFFF",
	"#FFA500",
	"#00FFFF",
	"#FFA500",
	"#00FFFF",
	"#FFA500",
	"#00FFFF",
	"#FFA500",
	"#00FFFF",
	"#FFA500",
	"#00FFFF",
}

type Vector2Int struct {
	X int
	Y int
}

func NewVector2Int(x, y int) Vector2Int {
	return Vector2Int{X: x, Y: y}
}

var landmarkEdges = []Vector2Int{
	NewVector2Int(15, 21),
	NewVector2Int(16, 20),
	NewVector2Int(18, 20),
	NewVector2Int(3, 7),
	NewVector2Int(14, 16),
	NewVector2Int(23, 25),
	NewVector2Int(28, 30),
	NewVector2Int(11, 23),
	NewVector2Int(27, 31),
	NewVector2Int(6, 8),
	NewVector2Int(15, 17),
	NewVector2Int(24, 26),
	NewVector2Int(16, 22),
	NewVector2Int(4, 5),
	NewVector2Int(5, 6),
	NewVector2Int(29, 31),
	NewVector2Int(12, 24),
	NewVector2Int(23, 24),
	NewVector2Int(0, 1),
	NewVector2Int(9, 10),
	NewVector2Int(1, 2),
	NewVector2Int(0, 4),
	NewVector2Int(11, 13),
	NewVector2Int(30, 32),
	NewVector2Int(28, 32),
	NewVector2Int(15, 19),
	NewVector2Int(16, 18),
	NewVector2Int(25, 27),
	NewVector2Int(26, 28),
	NewVector2Int(12, 14),
	NewVector2Int(17, 19),
	NewVector2Int(2, 3),
	NewVector2Int(11, 12),
	NewVector2Int(27, 29),
	NewVector2Int(13, 15),
}

type runningData struct {
	captures [][]position.Capture
}

func (rd *runningData) toRecording() format.Recording {
	childrenRecordings := make([]format.Recording, len(rd.captures))
	for i, col := range rd.captures {
		styling := metadata.EmptyBlock()
		styling.Mapping()["recolude-scale"] = metadata.NewStringProperty("0.04, 0.04, 0.04")
		styling.Mapping()["recolude-geom"] = metadata.NewStringProperty("sphere")
		styling.Mapping()["recolude-color"] = metadata.NewStringProperty(landmarkColors[i])

		childrenRecordings[i] = format.NewRecording(
			strconv.Itoa(i),
			landmarkNames[i],
			[]format.CaptureCollection{
				position.NewCollection("Position", col),
			},
			nil,
			styling,
			nil,
			nil,
		)
	}

	metadataLines := make([]metadata.Block, len(landmarkEdges))
	for i, link := range landmarkEdges {
		mapping := map[string]metadata.Property{
			"starting-object-id": metadata.NewStringProperty(strconv.Itoa(int(link.X))),
			"ending-object-id":   metadata.NewStringProperty(strconv.Itoa(int(link.Y))),
			"color":              metadata.NewStringProperty("#00FF00"),
			"width":              metadata.NewFloat32Property(0.01),
		}
		metadataLines[i] = metadata.NewBlock(mapping)
	}
	recordingMetadata := metadata.EmptyBlock()
	recordingMetadata.Mapping()["recolude-lines"] = metadata.NewMetadataArrayProperty(metadataLines)

	return format.NewRecording(
		"",
		"Pose Capture Demo",
		[]format.CaptureCollection{},
		childrenRecordings,
		recordingMetadata,
		nil,
		nil,
	)
}

func (rd *runningData) runDetection(curTime float64, frame []LandMark) {
	for i, landmark := range frame {
		if len(rd.captures) < i+1 {
			rd.captures = append(rd.captures, make([]position.Capture, 0))
		}
		rd.captures[i] = append(rd.captures[i], position.NewCapture(curTime, landmark.X*2, -landmark.Y*2, landmark.Z*2))
	}
}

type LandMark struct {
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	Z  float64 `json:"z"`
	ID int     `json:"id"`
}

func (lm LandMark) String() string {
	return fmt.Sprintf("%s: %f, %f, %f", landmarkNames[lm.ID], lm.X, lm.Y, lm.Z)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	jsonFile, err := os.Open("ivey.json")
	check(err)
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	check(err)

	// we initialize our Users array
	var frames [][]LandMark

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	check(json.Unmarshal(byteValue, &frames))

	rd := &runningData{
		captures: make([][]position.Capture, 0),
	}
	curTime := 0.0

	for _, frame := range frames {
		rd.runDetection(curTime, frame)
		curTime += 1.0 / 30.0
	}

	f, _ := os.Create("pose tracking.rap")
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
