package dds

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
)

// DXGI format
// https://learn.microsoft.com/en-us/windows/win32/api/dxgiformat/ne-dxgiformat-dxgi_format

const (
	DDSMagic = 0x20534444

	DXT1 = "DXT1"
	DXT3 = "DXT3"
	DXT5 = "DXT5"
	ATI1 = "ATI1"
	ATI2 = "ATI2"

  R32G32B32A32_TYPELESS = 1
  R32G32B32A32_FLOAT = 2
  R32G32B32A32_UINT = 3
  R32G32B32A32_SINT = 4
  R32G32B32_TYPELESS = 5
  R32G32B32_FLOAT = 6
  R32G32B32_UINT = 7
  R32G32B32_SINT = 8
  R16G16B16A16_TYPELESS = 9
  R16G16B16A16_FLOAT = 10
  R16G16B16A16_UNORM = 11
  R16G16B16A16_UINT = 12
  R16G16B16A16_SNORM = 13
  R16G16B16A16_SINT = 14
  R32G32_TYPELESS = 15
  R32G32_FLOAT = 16
  R32G32_UINT = 17
  R32G32_SINT = 18
  R32G8X24_TYPELESS = 19
  D32_FLOAT_S8X24_UINT = 20
  R32_FLOAT_X8X24_TYPELESS = 21
  X32_TYPELESS_G8X24_UINT = 22
  R10G10B10A2_TYPELESS = 23
  R10G10B10A2_UNORM = 24
  R10G10B10A2_UINT = 25
  R11G11B10_FLOAT = 26
  R8G8B8A8_TYPELESS = 27
  R8G8B8A8_UNORM = 28
  R8G8B8A8_UNORM_SRGB = 29
  R8G8B8A8_UINT = 30
  R8G8B8A8_SNORM = 31
  R8G8B8A8_SINT = 32
  R16G16_TYPELESS = 33
  R16G16_FLOAT = 34
  R16G16_UNORM = 35
  R16G16_UINT = 36
  R16G16_SNORM = 37
  R16G16_SINT = 38
  R32_TYPELESS = 39
  D32_FLOAT = 40
  R32_FLOAT = 41
  R32_UINT = 42
  R32_SINT = 43
  R24G8_TYPELESS = 44
  D24_UNORM_S8_UINT = 45
  R24_UNORM_X8_TYPELESS = 46
  X24_TYPELESS_G8_UINT = 47
  R8G8_TYPELESS = 48
  R8G8_UNORM = 49
  R8G8_UINT = 50
  R8G8_SNORM = 51
  R8G8_SINT = 52
  R16_TYPELESS = 53
  R16_FLOAT = 54
  D16_UNORM = 55
  R16_UNORM = 56
  R16_UINT = 57
  R16_SNORM = 58
  R16_SINT = 59
  R8_TYPELESS = 60
  R8_UNORM = 61
  R8_UINT = 62
  R8_SNORM = 63
  R8_SINT = 64
  A8_UNORM = 65
  R1_UNORM = 66
  R9G9B9E5_SHAREDEXP = 67
  R8G8_B8G8_UNORM = 68
  G8R8_G8B8_UNORM = 69
  BC1_TYPELESS = 70
  BC1_UNORM = 71
  BC1_UNORM_SRGB = 72
  BC2_TYPELESS = 73
  BC2_UNORM = 74
  BC2_UNORM_SRGB = 75
  BC3_TYPELESS = 76
  BC3_UNORM = 77
  BC3_UNORM_SRGB = 78
  BC4_TYPELESS = 79
  BC4_UNORM = 80
  BC4_SNORM = 81
  BC5_TYPELESS = 82
  BC5_UNORM = 83
  BC5_SNORM = 84
  B5G6R5_UNORM = 85
  B5G5R5A1_UNORM = 86
  B8G8R8A8_UNORM = 87
  B8G8R8X8_UNORM = 88
  R10G10B10_XR_BIAS_A2_UNORM = 89
  B8G8R8A8_TYPELESS = 90
  B8G8R8A8_UNORM_SRGB = 91
  B8G8R8X8_TYPELESS = 92
  B8G8R8X8_UNORM_SRGB = 93
  BC6H_TYPELESS = 94
  BC6H_UF16 = 95
  BC6H_SF16 = 96
  BC7_TYPELESS = 97
  BC7_UNORM = 98
  BC7_UNORM_SRGB = 99
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

type DX10Header struct {
	DXGIFormat        uint32
	ResourceDimension uint32
	MiscFlag          uint32
	ArraySize         uint32
	MiscFlags2        uint32
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

func ParseHeader(r io.Reader) (*DDHeader, *DX10Header, error) {
	var magic uint32
	if err := binary.Read(r, binary.LittleEndian, &magic); err != nil {
		return nil, nil, err
	}
	if magic != DDSMagic {
		return nil, nil, errors.New("not a valid DDS file")
	}

	header := &DDHeader{}
	if err := binary.Read(r, binary.LittleEndian, header); err != nil {
		fmt.Println("IF Generic == TRUE")
		return nil, nil, err
	}

	if header.Size != 124 {
		return nil, nil, errors.New("invalid DDS header size. Must be 124 bytes")
	}

	var dxt10 *DX10Header
	if string(header.Format.FourCC[:]) == "DX10" {
		dxt10 = &DX10Header{}
		if err := binary.Read(r, binary.LittleEndian, dxt10); err != nil {
			return nil, nil, err
		}
	}

	return header, dxt10, nil
}

func decodeBC1(r io.Reader, h *DDHeader) ([]color.RGBA, error) {
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

func decodeBC2(r io.Reader, h *DDHeader) ([]color.RGBA, error) {
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

func decodeBC3(r io.Reader, h *DDHeader) ([]color.RGBA, error) {
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

func decodeBC4(r io.Reader, h *DDHeader) ([]color.RGBA, error) {
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

func decodeBC5(r io.Reader, h *DDHeader) ([]color.RGBA, error) {
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
	h, dx10, err := ParseHeader(r)
	if err != nil {
		return nil, err
	}

	if dx10 != nil {
		switch dx10.DXGIFormat {
		case BC1_TYPELESS, BC1_UNORM, BC1_UNORM_SRGB:
			pix, err := decodeBC1(r, h)
			if err != nil {
				return nil, err
			}
			return &Image{Header: h, Pix: pix}, nil	
		case BC2_TYPELESS, BC2_UNORM, BC2_UNORM_SRGB:	
			pix, err := decodeBC2(r, h)
			if err != nil {
				return nil, err
			}
			return &Image{Header: h, Pix: pix}, nil
		case BC3_TYPELESS, BC3_UNORM, BC3_UNORM_SRGB:	
			pix, err := decodeBC3(r, h)
			if err != nil {
				return nil, err
			}
			return &Image{Header: h, Pix: pix}, nil
		case BC4_TYPELESS, BC4_UNORM, BC4_SNORM:
			pix, err := decodeBC4(r, h)
			if err != nil {
				return nil, err
			}
			return &Image{Header: h, Pix: pix}, nil
		case BC5_TYPELESS, BC5_UNORM, BC5_SNORM:
			pix, err := decodeBC5(r, h)
			if err != nil {
				return nil, err
			}
			return &Image{Header: h, Pix: pix}, nil
		default:
			dxgi_link := "https://learn.microsoft.com/en-us/windows/win32/api/dxgiformat/ne-dxgiformat-dxgi_format"
			return nil, fmt.Errorf("unsupported or unhandled DXGI format: [%d]\nSupported formats: %s", dx10.DXGIFormat, dxgi_link)
		}
	}

	fourCC := string(h.Format.FourCC[:])
	switch fourCC {
	case DXT1:
		pix, err := decodeBC1(r, h)
		if err != nil {
			return nil, err
		}
		return &Image{Header: h, Pix: pix}, nil
	case DXT3:
		pix, err := decodeBC2(r, h)
		if err != nil {
			return nil, err
		}
		return &Image{Header: h, Pix: pix}, nil
	case DXT5:
		pix, err := decodeBC3(r, h)
		if err != nil {
			return nil, err
		}
		return &Image{Header: h, Pix: pix}, nil
	case ATI1:
		pix, err := decodeBC4(r, h)
		if err != nil {
			return nil, err
		}
		return &Image{Header: h, Pix: pix}, nil
	case ATI2:
		pix, err := decodeBC5(r, h)
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
