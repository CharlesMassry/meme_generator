package main

import (
	"bytes"
	"flag"
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
	"github.com/satori/go.uuid"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var fontFile = loadFont("./font.ttf")

func main() {
	var PORT = flag.Int("PORT", 3001, "port to serve app on")

	flag.Parse()

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
		templateImage := loadImage(name)

		if templateImage != nil && (topText == "" || bottomText == "") {
			handleNotFound(reqID, req.URL.String(), w)
			return
		}

		log.Println(reqID, "Top Text:", topText, "Bottom Text:", bottomText)
		log.Println(reqID, name)

		go drawMemeText(templateImage, "top", topText)
		go drawMemeText(templateImage, "bottom", bottomText)

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
	log.Println("Listening on Port", strconv.Itoa(*PORT))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*PORT), nil))
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

func drawMemeText(img *image.RGBA, position string, text string) {
	offset := 75
	size := 42.0
	dpi := 72.0

	foreground := image.White

	h := font.HintingNone

	face := truetype.NewFace(fontFile, &truetype.Options{
		Size:    size,
		DPI:     dpi,
		Hinting: h,
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
		y = offset
	case "bottom":
		y = img.Bounds().Dy() - (offset - metrics.Ascent.Round() + metrics.Descent.Round())
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

func handleNotFound(reqID uuid.UUID, url string, w http.ResponseWriter) {
	log.Println(reqID, "404", url)
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 not found"))
}

func getText(req *http.Request) (string, string) {
	params := req.URL.Query()
	topText := convertToTitle(params.Get("top_text"))
	bottomText := convertToTitle(params.Get("bottom_text"))
	return topText, bottomText
}

func convertToTitle(str string) string {
	return strings.ToUpper(strings.Replace(str, "_", " ", -1))
}
