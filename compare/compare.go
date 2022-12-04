package compare

import (
	"apiFP/template/utils"
	"bytes"
	"encoding/binary"
	"math"
)

func ISO19794_BytesToTemplate(data []byte) utils.Template {
	reader := bytes.NewReader(data)

	var templateInp utils.TemplateInput
	var template utils.Template
	binary.Read(reader, binary.BigEndian, &templateInp.Header)
	binary.Read(reader, binary.BigEndian, &templateInp.Reserved)

	for i := 0; i < int(templateInp.Header.FpCount); i++ {
		var fp utils.FingerprintInput
		binary.Read(reader, binary.BigEndian, &fp.Position)
		binary.Read(reader, binary.BigEndian, &fp.ViewOffsetAndSampleType)
		binary.Read(reader, binary.BigEndian, &fp.FpQuality)
		binary.Read(reader, binary.BigEndian, &fp.MinutiaeCount)
		for j := 0; j < int(fp.MinutiaeCount); j++ {
			var m utils.MinutiaeInput
			binary.Read(reader, binary.BigEndian, &m.TypeAndXPos)
			binary.Read(reader, binary.BigEndian, &m.YPos)
			binary.Read(reader, binary.BigEndian, &m.Angle)
			binary.Read(reader, binary.BigEndian, &m.Quality)
			fp.Minutiae = append(fp.Minutiae, m)
		}
		binary.Read(reader, binary.BigEndian, &fp.ExtBytes)
		templateInp.Fingerprints = append(templateInp.Fingerprints, fp)
	}

	template.Header.Magic = string(templateInp.Header.Magic[:3])
	template.Header.Version = string(templateInp.Header.Version[:3])
	template.Header.TotalBytes = templateInp.Header.TotalBytes
	var dev uint8
	dev = templateInp.Header.Dev[0]*10 + templateInp.Header.Dev[1]
	template.Header.Dev = dev
	template.Header.Width = templateInp.Header.Width
	template.Header.Height = templateInp.Header.Height
	template.Header.ResX = templateInp.Header.ResX
	template.Header.ResY = templateInp.Header.ResY
	template.Header.FpCount = templateInp.Header.FpCount
	for _, fp := range templateInp.Fingerprints {
		template.Fingerprints = append(template.Fingerprints, utils.Fingerprint{
			Position:      fp.Position,
			ViewOffset:    fp.ViewOffsetAndSampleType & 0x0F,
			SampleType:    fp.ViewOffsetAndSampleType >> 4,
			FpQuality:     fp.FpQuality,
			MinutiaeCount: fp.MinutiaeCount,
			ExtBytes:      fp.ExtBytes,
		})
		for _, m := range fp.Minutiae {
			template.Fingerprints[len(template.Fingerprints)-1].Minutiae = append(template.Fingerprints[len(template.Fingerprints)-1].Minutiae, utils.Minutiae{
				Type:    m.TypeAndXPos >> 14,
				XPos:    m.TypeAndXPos & 0x0FFF,
				YPos:    m.YPos,
				Angle:   m.Angle,
				Quality: m.Quality,
			})
		}
	}
	return template
}

// getDistance returns the distance between two points
func getDistance(x0, y0, x, y uint16) int {
	dx := float64(x - x0)
	dy := float64(y - y0)
	return int(math.Sqrt(dx*dx + dy*dy))
}

// distance and angle checks. If the minutiae are within the specified distance and angle range, they are considered a match
func checkDistance(targetMinutiae, currentMinutiae utils.Minutiae, maxDistance int) bool {
	if getDistance(targetMinutiae.XPos, targetMinutiae.YPos, currentMinutiae.XPos, currentMinutiae.YPos) > maxDistance {
		return false
	}
	return true
}
func checkAngle(targetMinutiae, currentMinutiae utils.Minutiae, maxAngle int) bool {
	if math.Abs(float64(targetMinutiae.Angle-currentMinutiae.Angle)) > float64(maxAngle) {
		return false
	}
	return true
}

func Compare(target, current utils.Template) float64 {
	// Offset values. The lower the value, the more accurate the comparison
	var maxDistance = 15
	var maxAngle = 5

	var matched int

	// Loop through all target minutiae and compare them to the current minutiae
	for targetMinutiaeCount := 0; targetMinutiaeCount < int(target.Fingerprints[0].MinutiaeCount); targetMinutiaeCount++ {
		for currentMinutiaeCount := 0; currentMinutiaeCount < int(current.Fingerprints[0].MinutiaeCount); currentMinutiaeCount++ {
			if target.Fingerprints[0].Minutiae[targetMinutiaeCount].Type != current.Fingerprints[0].Minutiae[currentMinutiaeCount].Type {
				continue
			}
			if checkDistance(target.Fingerprints[0].Minutiae[targetMinutiaeCount], current.Fingerprints[0].Minutiae[currentMinutiaeCount], maxDistance) == false {
				continue
			}
			if checkAngle(target.Fingerprints[0].Minutiae[targetMinutiaeCount], current.Fingerprints[0].Minutiae[currentMinutiaeCount], maxAngle) == false {
				continue
			}
			matched++
		}
	}

	// Calculate the percentage of minutiae that matched. Related to the target minutiae count, NOT the current's
	percent := float64(matched) / float64(target.Fingerprints[0].MinutiaeCount)
	return percent
}
