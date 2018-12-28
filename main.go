package main

import (
	"bytes"
	"image"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var fontFile = loadFont("./font.ttf")

func main() {
	var PORT = os.Getenv("PORT")

	var TEMPLATES = map[string]bool{}

	files, err := ioutil.ReadDir("./memes")

	if err != nil {
		panic("couldn't read dir memes")
	}

	for _, fileInfo := range files {
		filename := fileInfo.Name()
		if strings.HasSuffix(filename, ".jpg") {
			memename := strings.TrimSuffix(filename, ".jpg")
			TEMPLATES[memename] = true
		}
	}

	if len(TEMPLATES) == 0 {
		panic("can't start server with no meme templates")
	}

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		reqID, err := uuid.NewV4()
		if err != nil {
			panic("error creating uuid")
		}

		log.Println(reqID, req.RemoteAddr)

		name := strings.TrimPrefix(req.URL.Path, "/")
		topText, bottomText := getText(req)
		_, ok := TEMPLATES[name]

		if ok == false && (topText == "" || bottomText == "") {
			handleNotFound(reqID, req.URL.String(), TEMPLATES, w)
			return
		}

		templateImage := loadImage(name)

		log.Println(reqID, "Top Text:", topText, "Bottom Text:", bottomText)
		log.Println(reqID, name)

		drawText(templateImage, topText, bottomText)

		jpegOptions := jpeg.Options{Quality: 65}
		var jpgBuffer bytes.Buffer
		jpeg.Encode(&jpgBuffer, templateImage, &jpegOptions)

		jpgBytes := jpgBuffer.Bytes()

		memeLength := len(jpgBytes)

		log.Println(reqID, "Generated Meme, length:", memeLength)
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Length", strconv.Itoa(memeLength))
		w.Write(jpgBytes)
	})
	log.Println("Listening on Port", PORT)
	log.Fatal(http.ListenAndServe(":"+PORT, nil))
}

func loadImage(name string) *image.RGBA {
	imageTemplateFile, err := os.Open("./memes/" + name + ".jpg")
	defer imageTemplateFile.Close()
	if err != nil {
		panic("error loading image")
	}

	templateImage, _, err := image.Decode(imageTemplateFile)

	if err != nil {
		panic("error decoding image")
	}

	bounds := templateImage.Bounds()
	rgbaImage := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))

	draw.Draw(rgbaImage, rgbaImage.Bounds(), templateImage, bounds.Min, draw.Src)

	return rgbaImage
}

func drawText(templateImage *image.RGBA, topText string, bottomText string) {

	// done := make(chan bool, 2)

	// go func(finished chan bool) {
	drawMemeText(templateImage, "top", topText)
	// finished <- true
	// }(done)

	// go func(finished chan bool) {
	drawMemeText(templateImage, "bottom", bottomText)
	// finished <- true
	// }(done)

	// <-done
}

type FontFaceOptions struct {
	Offset  int
	Size    float64
	DPI     float64
	Hinting font.Hinting
}

func drawMemeText(img *image.RGBA, position string, text string) {
	fontOptions := &FontFaceOptions{
		Offset:  75,
		Size:    42.0,
		DPI:     72.0,
		Hinting: font.HintingNone,
	}

	foreground := image.White

	face := truetype.NewFace(fontFile, &truetype.Options{
		Size:    fontOptions.Size,
		DPI:     fontOptions.DPI,
		Hinting: fontOptions.Hinting,
	})

	textBounds := 0

	for _, letter := range text {
		advance, ok := face.GlyphAdvance(letter)
		if ok {
			textBounds += advance.Round()
		}
	}

	metrics := face.Metrics()

	x := (img.Bounds().Max.X - textBounds) / 2

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
		Face: face,
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

func handleNotFound(reqID uuid.UUID, url string, templateNames map[string]bool, w http.ResponseWriter) {
	log.Println(reqID, "404", url)
	w.WriteHeader(http.StatusNotFound)
	responseString := "Valid Template Names:\n"

	for templateName := range templateNames {
		responseString += templateName + "\n"
	}

	w.Write([]byte(responseString))
}

func getText(req *http.Request) (string, string) {
	params := req.URL.Query()
	topText := convertToTitle(params.Get("top_text"))
	bottomText := convertToTitle(params.Get("bottom_text"))
	return topText, bottomText
}

func convertToTitle(str string) string {
	return strings.ToUpper(str)
}
