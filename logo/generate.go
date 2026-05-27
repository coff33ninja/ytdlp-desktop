package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"golang.org/x/image/draw"
)

var sizes = []int{16, 24, 32, 48, 64, 96, 128, 256}

func main() {
	srcPath := filepath.Join(".", "logo", "logo.png")
	outDir := filepath.Join(".", "logo")

	src, err := loadPNG(srcPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load %s: %v\n", srcPath, err)
		os.Exit(1)
	}

	for _, s := range sizes {
		out := resize(src, s)
		name := fmt.Sprintf("icon_%d.png", s)
		if err := savePNG(out, filepath.Join(outDir, name)); err != nil {
			fmt.Fprintf(os.Stderr, "save %s: %v\n", name, err)
		}
		fmt.Println("  ", name)
	}

	icoPath := filepath.Join(outDir, "app.ico")
	if err := makeICO(src, icoPath); err != nil {
		fmt.Fprintf(os.Stderr, "ico: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("  app.ico")

	svgPath := filepath.Join(outDir, "logo.svg")
	if err := makeSVGPlaceholder(svgPath); err != nil {
		fmt.Fprintf(os.Stderr, "svg: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("  logo.svg")
}

func loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

func savePNG(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func resize(src image.Image, size int) image.Image {
	dst := image.NewNRGBA(image.Rect(0, 0, size, size))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}

func makeICO(src image.Image, path string) error {
	ico := newICO()
	// embed 256px as PNG inside ICO (best quality)
	ico.addPNG(src, 256)
	// also embed 32px
	ico.addPNG(resize(src, 32), 32)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return ico.write(f)
}

type icoHeader struct {
	Reserved uint16
	Type     uint16
	Count    uint16
}

type icoEntry struct {
	Width   uint8
	Height  uint8
	Colors  uint8
	Reserve uint8
	Planes  uint16
	BPP     uint16
	Size    uint32
	Offset  uint32
}

type icoFile struct {
	entries []icoEntry
	data    [][]byte
}

func newICO() *icoFile {
	return &icoFile{}
}

func (i *icoFile) addPNG(img image.Image, size int) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return
	}
	w := uint8(0)
	h := uint8(0)
	if size < 256 {
		w = uint8(size)
		h = uint8(size)
	}
	i.entries = append(i.entries, icoEntry{
		Width:  w,
		Height: h,
		Planes: 1,
		BPP:    32,
		Size:   uint32(buf.Len()),
	})
	i.data = append(i.data, buf.Bytes())
}

func (i *icoFile) write(f *os.File) error {
	count := uint16(len(i.entries))
	hdr := icoHeader{Type: 1, Count: count}
	if err := binary.Write(f, binary.LittleEndian, &hdr); err != nil {
		return err
	}
	offset := uint32(6 + count*16)
	for idx := range i.entries {
		i.entries[idx].Offset = offset
		if err := binary.Write(f, binary.LittleEndian, &i.entries[idx]); err != nil {
			return err
		}
		offset += i.entries[idx].Size
	}
	for _, d := range i.data {
		if _, err := f.Write(d); err != nil {
			return err
		}
	}
	return nil
}

func makeSVGPlaceholder(path string) error {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256" viewBox="0 0 256 256">
  <rect width="256" height="256" rx="32" fill="#1a1a2e"/>
  <text x="128" y="140" font-family="sans-serif" font-size="64" font-weight="bold" fill="#66bbff" text-anchor="middle">YD</text>
</svg>`
	return os.WriteFile(path, []byte(svg), 0644)
}
