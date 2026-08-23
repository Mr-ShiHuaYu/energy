package winicon

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"io"

	"github.com/energye/energy/v2/cmd/internal/tools/winicon/internal/winicon"
	"golang.org/x/image/draw"
)

// xpCompatibleMaxSize XP 及以下系统不支持 PNG 压缩的图标条目(仅支持 BMP/DIB),
// 小于等于该尺寸的图标生成 BMP 格式条目, 大于该尺寸的生成 PNG 格式条目。
// XP 图标最大显示 48x48, 因此 XP 下使用 BMP 条目, Vista+ 使用 PNG 条目。
const xpCompatibleMaxSize = 48

// encodeIconBMP 将图像编码为 ICO 内使用的 BMP/DIB 格式条目数据
// 结构: BITMAPINFOHEADER(40字节) + XOR mask(BGRA 像素, 自底向上) + AND mask(1bpp, 行4字节对齐)
func encodeIconBMP(img *image.RGBA) []byte {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	buf := new(bytes.Buffer)
	// BITMAPINFOHEADER
	binary.Write(buf, binary.LittleEndian, uint32(40)) // biSize
	binary.Write(buf, binary.LittleEndian, int32(w))   // biWidth
	binary.Write(buf, binary.LittleEndian, int32(h*2)) // biHeight = XOR mask + AND mask 双倍高度
	binary.Write(buf, binary.LittleEndian, uint16(1))  // biPlanes
	binary.Write(buf, binary.LittleEndian, uint16(32)) // biBitCount
	binary.Write(buf, binary.LittleEndian, uint32(0))  // biCompression = BI_RGB
	binary.Write(buf, binary.LittleEndian, uint32(0))  // biSizeImage
	binary.Write(buf, binary.LittleEndian, int32(0))   // biXPelsPerMeter
	binary.Write(buf, binary.LittleEndian, int32(0))   // biYPelsPerMeter
	binary.Write(buf, binary.LittleEndian, uint32(0))  // biClrUsed
	binary.Write(buf, binary.LittleEndian, uint32(0))  // biClrImportant
	// XOR mask: BGRA 自底向上
	for y := h - 1; y >= 0; y-- {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			buf.WriteByte(byte(b >> 8))
			buf.WriteByte(byte(g >> 8))
			buf.WriteByte(byte(r >> 8))
			buf.WriteByte(byte(a >> 8))
		}
	}
	// AND mask: 1bpp, 每行按4字节对齐, 32bpp 透明度由 alpha 通道决定, 全部填 0
	andRowSize := ((w + 31) / 32) * 4
	buf.Write(make([]byte, andRowSize*h))
	return buf.Bytes()
}

// GenerateIcon reads image data from the given reader and generates
// a .ico file that is written to the given writer. The .ico file will include
// a number of icons at the sizes given.
func GenerateIcon(r io.Reader, w io.Writer, sizes []int) error {
	header := &winicon.IconFileHeader{
		ImageType:  1,
		ImageCount: uint16(len(sizes)),
	}

	iconheaders := make([]winicon.IconHeader, len(sizes))

	var imageData bytes.Buffer

	// Decode to internal image
	imagedata, _, err := image.Decode(r)
	if err != nil {
		return err
	}

	// Loop over sizes desired
	for index, size := range sizes {

		// Check target size
		if size == 0 {
			return fmt.Errorf("a size of 0 is not valid")
		}

		// Scale image
		rect := image.Rect(0, 0, size, size)
		rawdata := image.NewRGBA(rect)
		scale := draw.CatmullRom
		scale.Scale(rawdata, rect, imagedata, imagedata.Bounds(), draw.Over, nil)

		// 编码图标条目数据
		// 小尺寸使用 BMP/DIB 格式(XP 兼容), 大尺寸使用 PNG 格式(Vista+ 支持, 文件更小)
		var icondata []byte
		if size <= xpCompatibleMaxSize {
			icondata = encodeIconBMP(rawdata)
		} else {
			pngdata := new(bytes.Buffer)
			writer := bufio.NewWriter(pngdata)
			err = png.Encode(writer, rawdata)
			if err != nil {
				return err
			}
			err = writer.Flush()
			if err != nil {
				return err
			}
			icondata = pngdata.Bytes()
		}

		// Save image data
		imageData.Write(icondata)

		// Save header information
		if size >= 256 {
			size = 0
		}
		iconheaders[index].Width = (uint8)(size)
		iconheaders[index].Height = (uint8)(size)
		iconheaders[index].BitsPerPixel = 32
		iconheaders[index].Size = uint32(len(icondata))
	}

	// Update the offsets. Start by skipping header+icon headers
	var currentOffset uint32 = (uint32)(6 + (16 * len(iconheaders)))

	for index := range iconheaders {
		iconheaders[index].Offset = currentOffset
		currentOffset += iconheaders[index].Size
	}

	// Write out the header
	err = binary.Write(w, binary.LittleEndian, header)
	if err != nil {
		return err
	}

	// Write out the icon headers
	for _, iconheader := range iconheaders {
		err = binary.Write(w, binary.LittleEndian, iconheader)
		if err != nil {
			return err
		}
	}

	// Write out the image data
	_, err = w.Write(imageData.Bytes())
	if err != nil {
		return err
	}
	return nil
}
