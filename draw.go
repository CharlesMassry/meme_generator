package main

import (
	"image"
	"io/ioutil"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var fontFile = loadFont("./font.ttf")

type FontFaceOptions struct {
	Offset  int
	Size    float64
	DPI     float64
	Hinting font.Hinting
}

func drawText(templateImage *image.RGBA, topText string, bottomText string) {
	done := make(chan bool, 2)

	go func(finished chan bool) {
		drawMemeText(templateImage, "top", topText)
		finished <- true
	}(done)

	go func(finished chan bool) {
		drawMemeText(templateImage, "bottom", bottomText)
		finished <- true
	}(done)

	<-done
}

var fontOptions = &FontFaceOptions{
	Offset:  75,
	Size:    42.0,
	DPI:     72.0,
	Hinting: font.HintingNone,
}

func drawMemeText(img *image.RGBA, position string, text string) {
	var fontFace = truetype.NewFace(fontFile, &truetype.Options{
		Size:    fontOptions.Size,
		DPI:     fontOptions.DPI,
		Hinting: fontOptions.Hinting,
	})

	var foreground = image.White
	textBounds := 0

	for _, letter := range text {
		advance, ok := fontFace.GlyphAdvance(letter)
		if ok {
			textBounds += advance.Round()
		}
	}

	metrics := fontFace.Metrics()

	imageWidth := img.Bounds().Max.X

	// if imageWidth < textBounds {
	// splitLines(imageWidth, textBounds, text)
	// }

	x := (imageWidth - textBounds) / 2

	var y int
	switch position {
	case "top":
		y = fontOptions.Offset
	case "bottom":
		y = img.Bounds().Dy() - (fontOptions.Offset - metrics.Ascent.Round() + metrics.Descent.Round())
	default:
		panic("add Label function called without valid position")
	}

	drawer := &font.Drawer{
		Dst:  img,
		Src:  foreground,
		Face: fontFace,
	}

	drawer.Dot = fixed.P(x, y)

	drawer.DrawString(text)
}

func loadFont(fontfile string) *truetype.Font {
	fontBytes, err := ioutil.ReadFile(fontfile)
	if err != nil {
		panic("error reading the font file")
	}

	font, err := freetype.ParseFont(fontBytes)

	if err != nil {
		panic("error parsing the font file")
	}

	return font
}

// func splitLines(imageWidth int, textLength int, text string) []string {
// 	words := strings.Split(text, " ")

// 	amount := int(math.Ceil(float64(imageWidth) / float64(textLength)))

// 	len(words) / amount

// 	var lines []string

// 	return lines
// }
