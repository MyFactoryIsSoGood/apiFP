package compare

import (
	"apiFP/template/utils"
	"bytes"
	"encoding/binary"
	"math"
	//"encoding/json"
	"os"
)

//type Template struct {
//	Header       Header
//	Fingerprints []Fingerprint
//}
//
//type Header struct {
//	Magic      string
//	Version    string
//	TotalBytes uint32
//	Dev        uint8 // это я расшифровывать не стал, потому что нам оно не надо даже
//	Width      uint16
//	Height     uint16
//	ResX       uint16
//	ResY       uint16
//	FpCount    uint8
//}
//
//type Fingerprint struct {
//	Position      uint8
//	ViewOffset    uint8
//	SampleType    uint8
//	FpQuality     uint8
//	MinutiaeCount uint8
//	Minutiae      []Minutiae
//	ExtBytes      uint16
//}
//
//type Minutiae struct {
//	Type    uint16
//	XPos    uint16
//	YPos    uint16
//	Angle   uint8
//	Quality uint8
//}
//
//// Вспомогательные схемы
//type _TemplateInput struct {
//	Header       _HeaderInput
//	Reserved     uint8
//	Fingerprints []_FingerprintInput
//}
//
//type _HeaderInput struct {
//	Magic      [4]byte // FMR
//	Version    [4]byte // 020
//	TotalBytes uint32
//	Dev        [2]byte // Dev info
//	Width      uint16
//	Height     uint16
//	ResX       uint16
//	ResY       uint16
//	FpCount    uint8
//}
//
//type _FingerprintInput struct {
//	Position                uint8
//	ViewOffsetAndSampleType uint8
//	FpQuality               uint8
//	MinutiaeCount           uint8
//	Minutiae                []_MinutiaeInput
//	ExtBytes                uint16
//}
//
//type _MinutiaeInput struct {
//	TypeAndXPos uint16
//	YPos        uint16
//	Angle       uint8
//	Quality     uint8
//}

func ISO19794_BinToTemplate(filename string) utils.Template {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
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

func getDistance(x0, y0, x, y uint16) int {
	dx := float64(x - x0)
	dy := float64(y - y0)
	return int(math.Sqrt(dx*dx + dy*dy))
}

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
	var maxDistance = 5
	var maxAngle = 5

	var matched int

	for targetFingerprintsCount := 0; targetFingerprintsCount < int(target.Header.FpCount); targetFingerprintsCount++ {
		for targetMinutiaeCount := 0; targetMinutiaeCount < int(target.Fingerprints[0].MinutiaeCount); targetMinutiaeCount++ {
			for currentFingerprintsCount := 0; currentFingerprintsCount < int(current.Header.FpCount); currentFingerprintsCount++ {
				for currentMinutiaeCount := 0; currentMinutiaeCount < int(current.Fingerprints[0].MinutiaeCount); currentMinutiaeCount++ {
					if target.Fingerprints[targetFingerprintsCount].Minutiae[targetMinutiaeCount].Type != current.Fingerprints[currentFingerprintsCount].Minutiae[currentMinutiaeCount].Type {
						continue
					}
					if checkDistance(target.Fingerprints[targetFingerprintsCount].Minutiae[targetMinutiaeCount], current.Fingerprints[currentFingerprintsCount].Minutiae[currentMinutiaeCount], maxDistance) == false {
						continue
					}
					if checkAngle(target.Fingerprints[targetFingerprintsCount].Minutiae[targetMinutiaeCount], current.Fingerprints[currentFingerprintsCount].Minutiae[currentMinutiaeCount], maxAngle) == false {
						continue
					}
					matched++
				}
			}
		}
	}
	//println("Target:", target.Fingerprints[0].MinutiaeCount)
	//println("Current:", current.Fingerprints[0].MinutiaeCount)
	//println("Matched:", matched)

	percent := float64(matched) / float64(target.Fingerprints[0].MinutiaeCount)
	return percent
	//data, _ := json.MarshalIndent(templ, "", " ")
	//_ = os.WriteFile("1111.json", data, 0644)
}
