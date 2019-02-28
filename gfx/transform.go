package gfx

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/jeromelesaux/m4client/cpc"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
)

var (
	ColorNotFound     = errors.New("Color not found in palette")
	NotYetImplemented = errors.New("Function is not yet implemented")
)

func Transform(in *image.NRGBA, p color.Palette, size Size, filepath string) error {
	switch size {
	case Mode0:
		return TransformMode0(in, p, size, filepath)
	default:
		return NotYetImplemented
	}
	return nil
}

func PalettePosition(c color.Color, p color.Palette) (int, error) {
	r, g, b, a := c.RGBA()
	for index, cp := range p {
		//fmt.Fprintf(os.Stdout,"index(%d), c:%v,cp:%v\n",index,c,cp)
		rp, gp, bp, ap := cp.RGBA()
		if r == rp && g == gp && b == bp && a == ap {
			//fmt.Fprintf(os.Stdout,"Position found")
			return index, nil
		}
	}
	return -1, ColorNotFound
}

func TransformMode0(in *image.NRGBA, p color.Palette, size Size, filePath string) error {
	bw := make([]byte, 0)
	firmwareColorUsed := make(map[int]int,0)
	fmt.Fprintf(os.Stdout, "Informations palette (%d) for image (%d,%d)\n", len(p), in.Bounds().Max.X, in.Bounds().Max.Y)
	fmt.Println(in.Bounds())
	for j := 0; j < 8; j++ {
	for y := in.Bounds().Min.Y; y < in.Bounds().Max.Y/8; y ++ {
			for x := in.Bounds().Min.X; x+2 < in.Bounds().Max.X; x += 2 {
				c1 := in.At(x, y+j)
				pp1, err := PalettePosition(c1,p)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%v pixel position(%d,%d) not found in palette\n", c1, x, y+j)
					continue
				}
				firmwareColorUsed[pp1]++
				//fmt.Fprintf(os.Stdout, "(%d,%d), %v, position palette %d\n", x, y+j, c1, pp1)
				c2 := in.At(x+1, y+j)
				pp2, err := PalettePosition(c2,p)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%v pixel position(%d,%d) not found in paletqte\n", c2, x+1, y+j)
					continue
				}
				
				firmwareColorUsed[pp2]++
				
				//fmt.Fprintf(os.Stdout, "(%d,%d), %v, position palette %d\n", x+1, y+j, c1, pp2)
				var pixel byte
				//fmt.Fprintf(os.Stderr,"1:(%.8b)2:(%.8b)4:(%.8b)8:(%.8b)\n",1,2,4,8)
				//fmt.Fprintf(os.Stderr,"uint8(pp1)&1:%.8b\n",uint8(pp1)&1)
				if uint8(pp1)&1 == 1 {
					pixel += 128
				}
				//fmt.Fprintf(os.Stderr,"uint8(pp1)&2:%.8b\n",uint8(pp1)&2)
				if uint8(pp1)&2 == 2 {
					pixel += 8
				}
				//fmt.Fprintf(os.Stderr,"uint8(pp1)&4:%.8b\n",uint8(pp1)&4)
				if uint8(pp1)&4 == 4 {
					pixel += 32
				}
				//fmt.Fprintf(os.Stderr,"uint8(pp1)&8:%.8b\n",uint8(pp1)&8)
				if uint8(pp1)&8 == 8 {
					pixel += 2
				}
				if uint8(pp2)&1 == 1 {
					pixel += 64
				}
				if uint8(pp2)&2 == 2 {
					pixel += 4
				}
				if uint8(pp2)&4 == 4 {
					pixel += 16
				}
				if uint8(pp2)&8 == 8 {
					pixel++
				}
				fmt.Fprintf(os.Stderr, "pp1(%.8b), pp2(%.8b) pixel(%.8b)(%d)(&%.2x)\n", pp1, pp2, pixel, pixel, pixel)
				// MACRO PIXM0 COL2,COL1
				// ({COL1}&8)/8 | (({COL1}&4)*4) | (({COL1}&2)*2) | (({COL1}&1)*64) | (({COL2}&8)/4) | (({COL2}&4)*8) | (({COL2}&2)*4) | (({COL2}&1)*128)
				//	MEND
				//pixel = (uint8(pp1)&8)/8 | ((uint8(pp1)&4)*4) | ((uint8(pp1)&2)*2) | ((uint8(pp1)&1)*64) | ((uint8(pp2)&8)/4) | ((uint8(pp2)&4)*8) | ((uint8(pp2)&2)*4) | ((uint8(pp2)&1)*128)
				//pixel = (uint8(pp2) & 128)>>7  + (uint8(pp1) & 32)>>4  + (uint8(pp2) & 8)>>1 + (uint8(pp1) & 2)<<2 +
				// (uint8(pp2) & 64 )>>6 + (uint8(pp1) & 16)>>3  + (uint8(pp2) & 4) + (uint8(pp1) & 1)<<3
				bw = append(bw, pixel)
			}
		}
	}
	fmt.Println(firmwareColorUsed)
	header := cpc.CpcHead{Type: 2, User: 0, Address: 0x4000, Exec: 0xC000, Size: 0x4000, Size2: 0x4000}
	filename := filepath.Base(filePath)
	extension := filepath.Ext(filename)
	cpcFilename := strings.ToUpper(strings.Replace(filename, extension, ".SCR", -1))
	copy(header.Filename[:], cpcFilename)
	header.Checksum = uint16(header.ComputedChecksum16())
	fmt.Fprintf(os.Stderr, "Header lenght %d", binary.Size(header))
	fw, err := os.Create(cpcFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while creating file (%s) error :%s", cpcFilename, err)
		return err
	}
	binary.Write(fw, binary.LittleEndian, header)
	binary.Write(fw, binary.LittleEndian, bw)
	//fw.Write(bw)
	fw.Close()

	return nil
}
