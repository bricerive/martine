package gfx

import (
	"os"
	"testing"
)

func TestSaveDelta(t *testing.T) {
	d := NewDeltaCollection()
	for i := 0; i < 320; i++ {
		d.Add(0xFF, uint16(i))
	}
	if err := d.Save("delta.bin"); err != nil {
		t.Fatalf("expected no error and gets %v\n", err)
	}
	filesize := 4 + (320 * 2)

	fi, err := os.Lstat("delta.bin")
	if err != nil {
		t.Fatalf("expected no error while getting informations gets :%v\n", err)
	}

	if fi.Size() != int64(filesize) {
		t.Fatalf("expected %d length and gets %d\n", filesize, fi.Size())
	}
	//os.Remove("delta.bin")
}
