package main

import (
	"image"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core"

	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/line"

	"github.com/johnfercher/maroto/v2"

	"github.com/johnfercher/maroto/v2/pkg/components/code"
	img "github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/signature"
	"github.com/johnfercher/maroto/v2/pkg/components/text"

	"image/jpeg"
	_ "image/jpeg"

	"github.com/nfnt/resize"
)

func main() {
	m := GetMaroto()
	document, err := m.Generate()
	if err != nil {
		log.Fatal(err)
	}

	err = document.Save("./examples/pdf/simplestv2.pdf")
	if err != nil {
		log.Fatal(err)
	}
}

func GetMaroto() core.Maroto {
	cfg := config.NewBuilder().
		WithPageSize(pagesize.A6).
		Build()

	m := maroto.New(cfg)

	m.AddRow(20,
		code.NewQrCol(4, "https://github.com/adcondev/go-examples"),
	)

	m.AddRow(10, col.New(12))

	// Load and resize the image
	path := "./examples/pdf/gopher.jpg"
	resizedImg := loadAndResizeImage(path, 384) // Resize to width 384px (adjust as needed)

	m.AddRow(
		20,
		img.NewFromFileCol(4, resizedImg),
	)

	m.AddRow(
		20,
		signature.NewCol(4, "signature"),
	)

	m.AddRow(
		20,
		text.NewCol(4, "text"),
	)

	m.AddRow(20, line.NewCol(12))

	return m
}

// Helper function to load, resize, and save image
func loadAndResizeImage(filePath string, width uint) string {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	// Resize to specified width, maintaining aspect ratio
	resized := resize.Resize(width, 0, img, resize.Lanczos3)

	// Create output path (e.g., add "_resized" suffix)
	outputPath := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + "_resized" + filepath.Ext(filePath)

	// Save resized image as JPEG (adjust quality or format as needed)
	outFile, err := os.Create(outputPath)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	err = jpeg.Encode(outFile, resized, &jpeg.Options{Quality: 85}) // Quality 85 for balance
	if err != nil {
		log.Fatal(err)
	}

	return outputPath
}
