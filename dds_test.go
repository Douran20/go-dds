package dds

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestHeader(t *testing.T) {
	err := filepath.WalkDir("./resource", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		header, _, err := ParseHeader(file)
		if err != nil {
			t.Errorf("failed to parse %s: %v", path, err)
			return nil
		}
	
		fmt.Printf("\nFile = %s\n", path)
		
		fmt.Printf("Size = %d\n", header.Size)
		fmt.Printf("Flags = %d\n", header.Flags)
		fmt.Printf("Height = %d\n", header.Height)
		fmt.Printf("Width = %d\n", header.Width)
		fmt.Printf("Pitch = %d\n", header.PitchOrLinearSize)
		fmt.Printf("Depth = %d\n", header.Depth)
		fmt.Printf("MipMapCount = %d\n", header.MipMapCount)
		fmt.Printf("Reserved1 = %d\n", header.Reserved1[:])
		fmt.Printf("Format = %d\n", header.Format)
		fmt.Printf("Caps = %d\n", header.Caps)
		fmt.Printf("Caps2 = %d\n", header.Caps2)
		fmt.Printf("Caps3 = %d\n", header.Caps3)
		fmt.Printf("Caps4 = %d\n", header.Caps4)
		fmt.Printf("Reserved2 = %d\n", header.Reserved2)

		fmt.Printf("Comperssion Method %s\n", string(header.Format.FourCC[:]))
		
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestDXT10Header(t *testing.T) {
	err := filepath.WalkDir("./resource", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		header, DX10, err := ParseHeader(file)
		if err != nil {
			t.Errorf("failed to parse %s: %v", path, err)
			return nil
		}
		if string(header.Format.FourCC[:]) == "DX10" {
			fmt.Printf("\nFile = %s\n", path)
			fmt.Printf("Test = %d\n", DX10.DXGIFormat)
		}
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

