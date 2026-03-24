package dds 

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
)

const (
	DDSMagic = 0x20534444
	RawSize  = 124
)

const (
	DDPF_ALPHAPIXELS = 0x1
	DDPF_ALPHA       = 0x2
	DDPF_FOURCC      = 0x4
	DDPF_RGB         = 0x40
	DDPF_YUV         = 0x200
	DDPF_LUMINANCE   = 0x20000
)

type DDPixelFormat struct {
	Size        uint32
	Flags       uint32
	FourCC      [4]byte
	RGBBitCount uint32
	RbitMask    uint32
	GbitMask    uint32
	BbitMask    uint32
	AbitMask    uint32
}

type DDHeader struct {
	Size              uint32
	Flags             uint32
	Height            uint32
	Width             uint32
	PitchOrLinearSize uint32
	Depth             uint32
	MipMapCount       uint32
	Reserved1         [11]uint32
	Format            DDPixelFormat
	Caps              uint32
	Caps2             uint32
	Caps3             uint32
	Caps4             uint32
	Reserved2         uint32
}

type Image struct {
	Header *DDHeader
	Pix    []color.RGBA
}

func unpack565(val uint16) color.RGBA {
	r := uint8((val >> 11) & 0x1F)
	g := uint8((val >> 5) & 0x3F)
	b := uint8(val & 0x1F)
	return color.RGBA{
		R: (r << 3) | (r >> 2),
		G: (g << 2) | (g >> 4),
		B: (b << 3) | (b >> 2),
		A: 255,
	}
}

func lerp(c1, c2 color.RGBA, w1, w2 int) color.RGBA {
	sum := w1 + w2

	return color.RGBA{
		R: uint8((int(c1.R)*w1 + int(c2.R)*w2) / sum),
		G: uint8((int(c1.G)*w1 + int(c2.G)*w2) / sum),
		B: uint8((int(c1.B)*w1 + int(c2.B)*w2) / sum),
		A: 255,
	}
}

func ParseHeader(r io.Reader) (*DDHeader, error) {
	var magic uint32
	if err := binary.Read(r, binary.LittleEndian, &magic); err != nil {
		return nil, err
	}
	if magic != DDSMagic {
		return nil, errors.New("not a valid DDS file")
	}

	header := &DDHeader{}
	if err := binary.Read(r, binary.LittleEndian, header); err != nil {
		return nil, err
	}

	if header.Size != RawSize {
		return nil, errors.New("invalid DDS header size")
	}

	return header, nil
}

func decodeDXT1(r io.Reader, h *DDHeader) ([]color.RGBA, error) {
	width, height := int(h.Width), int(h.Height)
	pixels := make([]color.RGBA, width*height)

	for yBlock := 0; yBlock < height; yBlock += 4 {
		for xBlock := 0; xBlock < width; xBlock += 4 {

			var color0Raw, color1Raw uint16
			binary.Read(r, binary.LittleEndian, &color0Raw)
			binary.Read(r, binary.LittleEndian, &color1Raw)

			var indices uint32
			binary.Read(r, binary.LittleEndian, &indices)

			c0 := unpack565(color0Raw)
			c1 := unpack565(color1Raw)

			palette := [4]color.RGBA{c0, c1, {}, {}}
			if color0Raw > color1Raw {
				palette[2] = lerp(c0, c1, 2, 1) 
				palette[3] = lerp(c0, c1, 1, 2) 
			} else {
				palette[2] = lerp(c0, c1, 1, 1)     
				palette[3] = color.RGBA{0, 0, 0, 0} 
			}
	
			for py := 0; py < 4; py++ {
				for px := 0; px < 4; px++ {
					if xBlock+px < width && yBlock+py < height {	
						bitOffset := uint((py*4 + px) * 2)
						idx := (indices >> bitOffset) & 0x03

						pixelIdx := (yBlock+py)*width + (xBlock + px)
						pixels[pixelIdx] = palette[idx]
					}
				}
			}
		}
	}
	return pixels, nil
}

func decodeDXT3(r io.Reader, h *DDHeader) ([]color.RGBA, error) {
	width, height := int(h.Width), int(h.Height)
	pixels := make([]color.RGBA, width*height)

	for yBlock := 0; yBlock < height; yBlock += 4 {
		for xBlock := 0; xBlock < width; xBlock += 4 {
			
			var alphaRaw uint64
			binary.Read(r, binary.LittleEndian, &alphaRaw)

			var color0Raw, color1Raw uint16
			binary.Read(r, binary.LittleEndian, &color0Raw)
			binary.Read(r, binary.LittleEndian, &color1Raw)

			var cIndices uint32
			binary.Read(r, binary.LittleEndian, &cIndices)

			c0 := unpack565(color0Raw)
			c1 := unpack565(color1Raw)

			colorPalette := [4]color.RGBA{
				c0,
				c1,
				lerp(c0, c1, 2, 1),
				lerp(c0, c1, 1, 2),
			}

			for py := 0; py < 4; py++ {
				for px := 0; px < 4; px++ {
					if xBlock+px < width && yBlock+py < height {
						pixelNum := uint(py*4 + px)

						cIdx := (cIndices >> (pixelNum * 2)) & 0x03
						finalColor := colorPalette[cIdx]

						a4 := uint8((alphaRaw >> (pixelNum * 4)) & 0x0F)
						
						finalColor.A = (a4 << 4) | a4

						pixelIdx := (yBlock+py)*width + (xBlock + px)
						pixels[pixelIdx] = finalColor
					}
				}
			}
		}
	}
	return pixels, nil
}

func decodeDXT5(r io.Reader, h *DDHeader) ([]color.RGBA, error) {
	width, height := int(h.Width), int(h.Height)
	pixels := make([]color.RGBA, width*height)

	for yBlock := 0; yBlock < height; yBlock += 4 {
		for xBlock := 0; xBlock < width; xBlock += 4 {
			var a0, a1 uint8
			binary.Read(r, binary.LittleEndian, &a0)
			binary.Read(r, binary.LittleEndian, &a1)

			aIndicesRaw := make([]byte, 6)
			r.Read(aIndicesRaw)
			var aIndices uint64
			for i, b := range aIndicesRaw {
				aIndices |= uint64(b) << (8 * i)
			}

			alphaPalette := [8]uint8{}
			alphaPalette[0] = a0
			alphaPalette[1] = a1
			if a0 > a1 {
				for i := 1; i < 7; i++ {
					alphaPalette[i+1] = uint8((float64(7-i)*float64(a0) + float64(i)*float64(a1)) / 7.0)
				}
			} else {
				for i := 1; i < 5; i++ {
					alphaPalette[i+1] = uint8((float64(5-i)*float64(a0) + float64(i)*float64(a1)) / 5.0)
				}
				alphaPalette[6] = 0
				alphaPalette[7] = 255
			}

			var color0Raw, color1Raw uint16
			binary.Read(r, binary.LittleEndian, &color0Raw)
			binary.Read(r, binary.LittleEndian, &color1Raw)

			var cIndices uint32
			binary.Read(r, binary.LittleEndian, &cIndices)

			c0 := unpack565(color0Raw)
			c1 := unpack565(color1Raw)

			colorPalette := [4]color.RGBA{
				c0,
				c1,
				lerp(c0, c1, 2, 1),
				lerp(c0, c1, 1, 2),
			}

			for py := 0; py < 4; py++ {
				for px := 0; px < 4; px++ {
					if xBlock+px < width && yBlock+py < height {
						pixelNum := uint(py*4 + px)

						cIdx := (cIndices >> (pixelNum * 2)) & 0x03
						finalColor := colorPalette[cIdx]

						aIdx := (aIndices >> (pixelNum * 3)) & 0x07
						finalColor.A = alphaPalette[aIdx]

						pixelIdx := (yBlock+py)*width + (xBlock + px)
						pixels[pixelIdx] = finalColor
					}
				}
			}
		}
	}
	return pixels, nil
}

func readBC4Block(r io.Reader) ([8]uint8, uint64) {
	var v0, v1 uint8
	binary.Read(r, binary.LittleEndian, &v0)
	binary.Read(r, binary.LittleEndian, &v1)

	idxRaw := make([]byte, 6)
	r.Read(idxRaw)
	var indices uint64
	for i, b := range idxRaw {
		indices |= uint64(b) << (8 * i)
	}

	palette := [8]uint8{v0, v1}
	if v0 > v1 {
		for i := 1; i < 7; i++ {
			palette[i+1] = uint8((int(7-i)*int(v0) + int(i)*int(v1)) / 7)
		}
	} else {
		for i := 1; i < 5; i++ {
			palette[i+1] = uint8((int(5-i)*int(v0) + int(i)*int(v1)) / 5)
		}
		palette[6] = 0
		palette[7] = 255
	}
	return palette, indices
}

func decodeATI1(r io.Reader, h *DDHeader) ([]color.RGBA, error) {
	width, height := int(h.Width), int(h.Height)
	pixels := make([]color.RGBA, width*height)

	for yBlock := 0; yBlock < height; yBlock += 4 {
		for xBlock := 0; xBlock < width; xBlock += 4 {
			palette, indices := readBC4Block(r)

			for py := 0; py < 4; py++ {
				for px := 0; px < 4; px++ {
					if xBlock+px < width && yBlock+py < height {
						pixelNum := uint(py*4 + px)
						idx := (indices >> (pixelNum * 3)) & 0x07
						val := palette[idx]

						pixelIdx := (yBlock+py)*width + (xBlock + px)
						pixels[pixelIdx] = color.RGBA{val, val, val, 255}
					}
				}
			}
		}
	}
	return pixels, nil
}

func decodeATI2(r io.Reader, h *DDHeader) ([]color.RGBA, error) {
	width, height := int(h.Width), int(h.Height)
	pixels := make([]color.RGBA, width*height)

	for yBlock := 0; yBlock < height; yBlock += 4 {
		for xBlock := 0; xBlock < width; xBlock += 4 {
			
			redPalette, redIndices := readBC4Block(r)

			greenPalette, greenIndices := readBC4Block(r)

			for py := 0; py < 4; py++ {
				for px := 0; px < 4; px++ {
					if xBlock+px < width && yBlock+py < height {
						pixelNum := uint(py*4 + px)
						
						rIdx := (redIndices >> (pixelNum * 3)) & 0x07
						gIdx := (greenIndices >> (pixelNum * 3)) & 0x07

						pixelIdx := (yBlock+py)*width + (xBlock + px)
						pixels[pixelIdx] = color.RGBA{
							R: redPalette[rIdx],
							G: greenPalette[gIdx],
							B: 0, // need to check if BC5 actually stores a Z component 
							A: 255,
						}
					}
				}
			}
		}
	}
	return pixels, nil
}

func init() {
	image.RegisterFormat("dds", "DDS ", Decode, nil)
}

func Decode(r io.Reader) (image.Image, error) {
	h, err := ParseHeader(r)
	if err != nil {
		return nil, err
	}

	fourCC := string(h.Format.FourCC[:])

	switch fourCC {
	case "DXT1":
		pix, err := decodeDXT1(r, h)
		if err != nil {
			return nil, err
		}
		return &Image{Header: h, Pix: pix}, nil
	case "DXT3":
		pix, err := decodeDXT3(r, h)
		if err != nil {
			return nil, err
		}
		return &Image{Header: h, Pix: pix}, nil
	case "DXT5":
		pix, err := decodeDXT5(r, h)
		if err != nil {
			return nil, err
		}
		return &Image{Header: h, Pix: pix}, nil
	case "ATI1":
		pix, err := decodeATI1(r, h)
		if err != nil {
			return nil, err
		}
		return &Image{Header: h, Pix: pix}, nil
	case "ATI2":
		pix, err := decodeATI2(r, h)
		if err != nil {
			return nil, err
		}
		return &Image{Header: h, Pix: pix}, nil
	default:
		return nil, fmt.Errorf("unsupported or unhandled DDS format: [%s]", fourCC)
	}
}

func (i *Image) ColorModel() color.Model { return color.RGBAModel }
func (i *Image) Bounds() image.Rectangle {
	return image.Rect(0, 0, int(i.Header.Width), int(i.Header.Height))
}
func (i *Image) At(x, y int) color.Color {
	return i.Pix[y*int(i.Header.Width)+x]
}
