package utils

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

const (
	Pore MinType = iota
	Bifurcation
	Termination
	Unknown
)

type MinType byte

type Template struct {
	Header       Header
	Fingerprints []Fingerprint
}

type Header struct {
	Magic      string
	Version    string
	TotalBytes uint32
	Dev        uint8
	Width      uint16
	Height     uint16
	ResX       uint16
	ResY       uint16
	FpCount    uint8
}

type Fingerprint struct {
	Position      uint8
	ViewOffset    uint8
	SampleType    uint8
	FpQuality     uint8
	MinutiaeCount uint8
	Minutiae      []Minutiae
	ExtBytes      uint16
}

type Minutiae struct {
	Type    uint16
	XPos    uint16
	YPos    uint16
	Angle   uint8
	Quality uint8
}

// Вспомогательные схемы
type TemplateInput struct {
	Header       HeaderInput
	Reserved     uint8
	Fingerprints []FingerprintInput
}

type HeaderInput struct {
	Magic      [4]byte // FMR
	Version    [4]byte // 020
	TotalBytes uint32
	Dev        [2]byte // Dev info
	Width      uint16
	Height     uint16
	ResX       uint16
	ResY       uint16
	FpCount    uint8
}

type FingerprintInput struct {
	Position                uint8
	ViewOffsetAndSampleType uint8
	FpQuality               uint8
	MinutiaeCount           uint8
	Minutiae                []MinutiaeInput
	ExtBytes                uint16
}

type MinutiaeInput struct {
	TypeAndXPos uint16
	YPos        uint16
	Angle       uint8
	Quality     uint8
}

func ISO19794_BinToTemplate(filename string) Template {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	reader := bytes.NewReader(data)

	var templateInp TemplateInput
	var template Template
	binary.Read(reader, binary.BigEndian, &templateInp.Header)
	binary.Read(reader, binary.BigEndian, &templateInp.Reserved)

	for i := 0; i < int(templateInp.Header.FpCount); i++ {
		var fp FingerprintInput
		binary.Read(reader, binary.BigEndian, &fp.Position)
		binary.Read(reader, binary.BigEndian, &fp.ViewOffsetAndSampleType)
		binary.Read(reader, binary.BigEndian, &fp.FpQuality)
		binary.Read(reader, binary.BigEndian, &fp.MinutiaeCount)
		for j := 0; j < int(fp.MinutiaeCount); j++ {
			var m MinutiaeInput
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
		template.Fingerprints = append(template.Fingerprints, Fingerprint{
			Position:      fp.Position,
			ViewOffset:    fp.ViewOffsetAndSampleType & 0x0F,
			SampleType:    fp.ViewOffsetAndSampleType >> 4,
			FpQuality:     fp.FpQuality,
			MinutiaeCount: fp.MinutiaeCount,
			ExtBytes:      fp.ExtBytes,
		})
		for _, m := range fp.Minutiae {
			template.Fingerprints[len(template.Fingerprints)-1].Minutiae = append(template.Fingerprints[len(template.Fingerprints)-1].Minutiae, Minutiae{
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

func getViewOffsetAndSampleType(viewOffsetInput, sampleTypeInput uint8) uint8 {
	var sampleTypePart uint8 = (sampleTypeInput << 4) & 0xF0
	return sampleTypePart + viewOffsetInput
}

func getMinutiaeTypeAndXPos(typeInput, XPosInput uint16) uint16 {
	var TypePart uint16 = (typeInput << 14)
	return TypePart + XPosInput
}

func ISO19794TemplateToBin(templateStructure Template, outputFileName string) {
	outputFileHandle, err := os.Create(outputFileName)
	if err != nil {
		println("file creation error")
		os.Exit(0)
	}
	WriteBuffer := bufio.NewWriter(outputFileHandle)
	WriteBuffer.WriteString(templateStructure.Header.Magic)
	WriteBuffer.Flush()
	WriteBuffer.WriteByte(0x00)
	WriteBuffer.WriteString(templateStructure.Header.Version)
	WriteBuffer.WriteByte(0x00)
	binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Header.TotalBytes)
	binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Header.Dev)
	WriteBuffer.WriteByte(0x00)
	WriteBuffer.Flush()
	binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Header.Width)
	binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Header.Height)
	WriteBuffer.Flush()
	binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Header.ResX)
	binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Header.ResY)
	binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Header.FpCount)
	WriteBuffer.WriteByte(0x00)
	WriteBuffer.Flush()
	for i := 0; i < int(templateStructure.Header.FpCount); i++ {
		binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Fingerprints[i].Position)
		var viewOffsetOutput = getViewOffsetAndSampleType(templateStructure.Fingerprints[i].ViewOffset, templateStructure.Fingerprints[i].SampleType)
		binary.Write(WriteBuffer, binary.BigEndian, &viewOffsetOutput)
		binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Fingerprints[i].FpQuality)
		binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Fingerprints[i].MinutiaeCount)
		WriteBuffer.Flush()
		for j := 0; j < int(templateStructure.Fingerprints[i].MinutiaeCount); j++ {
			var XPosAndTypeOutput = getMinutiaeTypeAndXPos(templateStructure.Fingerprints[i].Minutiae[j].Type, templateStructure.Fingerprints[i].Minutiae[j].XPos)
			binary.Write(WriteBuffer, binary.BigEndian, &XPosAndTypeOutput)
			binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Fingerprints[i].Minutiae[j].YPos)
			binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Fingerprints[i].Minutiae[j].Angle)
			binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Fingerprints[i].Minutiae[j].Quality)
			WriteBuffer.Flush()
		}
	}
	WriteBuffer.WriteByte(0x00)
	WriteBuffer.WriteByte(0x00)
	WriteBuffer.Flush()
}

func ISO19794TemplateToBytes(templateStructure Template) []byte {
	//outputFileHandle, err := os.Create(outputFileName)
	//if err != nil {
	//	println("file creation error")
	//	os.Exit(0)
	//}
	resp := new(bytes.Buffer)
	WriteBuffer := bufio.NewWriter(resp)
	WriteBuffer.WriteString(templateStructure.Header.Magic)
	WriteBuffer.Flush()
	WriteBuffer.WriteByte(0x00)
	WriteBuffer.WriteString(templateStructure.Header.Version)
	WriteBuffer.WriteByte(0x00)
	binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Header.TotalBytes)
	binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Header.Dev)
	WriteBuffer.WriteByte(0x00)
	WriteBuffer.Flush()
	binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Header.Width)
	binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Header.Height)
	WriteBuffer.Flush()
	binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Header.ResX)
	binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Header.ResY)
	binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Header.FpCount)
	WriteBuffer.WriteByte(0x00)
	WriteBuffer.Flush()
	for i := 0; i < int(templateStructure.Header.FpCount); i++ {
		binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Fingerprints[i].Position)
		var viewOffsetOutput = getViewOffsetAndSampleType(templateStructure.Fingerprints[i].ViewOffset, templateStructure.Fingerprints[i].SampleType)
		binary.Write(WriteBuffer, binary.BigEndian, &viewOffsetOutput)
		binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Fingerprints[i].FpQuality)
		binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Fingerprints[i].MinutiaeCount)
		WriteBuffer.Flush()
		for j := 0; j < int(templateStructure.Fingerprints[i].MinutiaeCount); j++ {
			var XPosAndTypeOutput = getMinutiaeTypeAndXPos(templateStructure.Fingerprints[i].Minutiae[j].Type, templateStructure.Fingerprints[i].Minutiae[j].XPos)
			binary.Write(WriteBuffer, binary.BigEndian, &XPosAndTypeOutput)
			binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Fingerprints[i].Minutiae[j].YPos)
			binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Fingerprints[i].Minutiae[j].Angle)
			binary.Write(WriteBuffer, binary.BigEndian, &templateStructure.Fingerprints[i].Minutiae[j].Quality)
			WriteBuffer.Flush()
		}
	}
	WriteBuffer.WriteByte(0x00)
	WriteBuffer.WriteByte(0x00)
	WriteBuffer.Flush()
	return resp.Bytes()
}

func ISO19794_BytesToTemplate(data []byte) Template {
	reader := bytes.NewReader(data)

	var templateInp TemplateInput
	var template Template
	binary.Read(reader, binary.BigEndian, &templateInp.Header)
	binary.Read(reader, binary.BigEndian, &templateInp.Reserved)

	for i := 0; i < int(templateInp.Header.FpCount); i++ {
		var fp FingerprintInput
		binary.Read(reader, binary.BigEndian, &fp.Position)
		binary.Read(reader, binary.BigEndian, &fp.ViewOffsetAndSampleType)
		binary.Read(reader, binary.BigEndian, &fp.FpQuality)
		binary.Read(reader, binary.BigEndian, &fp.MinutiaeCount)
		for j := 0; j < int(fp.MinutiaeCount); j++ {
			var m MinutiaeInput
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
		template.Fingerprints = append(template.Fingerprints, Fingerprint{
			Position:      fp.Position,
			ViewOffset:    fp.ViewOffsetAndSampleType & 0x0F,
			SampleType:    fp.ViewOffsetAndSampleType >> 4,
			FpQuality:     fp.FpQuality,
			MinutiaeCount: fp.MinutiaeCount,
			ExtBytes:      fp.ExtBytes,
		})
		for _, m := range fp.Minutiae {
			template.Fingerprints[len(template.Fingerprints)-1].Minutiae = append(template.Fingerprints[len(template.Fingerprints)-1].Minutiae, Minutiae{
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

func SavePNGImage(in *image.Gray, filename string) {
	f, _ := os.Create(filename)
	png.Encode(f, in)
	f.Close()
}

func SavePNGImageColored(in *image.RGBA, filename string) {
	f, _ := os.Create(filename)
	png.Encode(f, in)
	f.Close()
}

func SetRectOnImg(in *image.RGBA, col color.RGBA, xPos, yPos int) {
	var width = 5
	for y := yPos - width; y < yPos+width; y++ {
		if y == yPos-width || y == yPos+width-1 {
			for x := xPos - width; x < xPos+width; x++ {
				in.Set(x, y, col)
			}
		} else {
			in.Set(xPos-width, y, col)
			in.Set(xPos+width-1, y, col)
		}
	}
}

func DisplayMinutiae(list []Minutiae, in *image.Gray) *image.RGBA {
	out := image.NewRGBA(in.Bounds())
	draw.Draw(out, in.Bounds(), in, image.Point{}, draw.Src)
	for _, minutiae := range list {
		if minutiae.Type == uint16(Termination) {
			SetRectOnImg(out, color.RGBA{R: 255, G: 0, B: 0, A: 255}, int(minutiae.XPos), int(minutiae.YPos))
		} else if minutiae.Type == uint16(Bifurcation) {
			SetRectOnImg(out, color.RGBA{R: 0, G: 0, B: 255, A: 255}, int(minutiae.XPos), int(minutiae.YPos))
		}
	}
	return out
}
