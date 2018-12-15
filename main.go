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

	"github.com/satori/go.uuid"
)

const PORT int = 8080

var TEMPLATES = map[string]bool{
	"roll_safe":     true,
	"scumbag_steve": true,
}

var fontfile string = "./impact.ttf"

func main() {
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		reqID, err := uuid.NewV4()
		if err != nil {
			panic("error creating uuid")
		}

		name := strings.TrimPrefix(req.URL.Path, "/")
		topText, bottomText := getText(req)

		if !TEMPLATES[name] && topText == "" && bottomText == "" {
			handleNotFound(reqID, req.URL.String(), w)
			return
		}

		log.Println(reqID, "Top Text:", topText, "Bottom Text:", bottomText)
		log.Println(reqID, name)

		templateImage := loadImage(name)
		// font := loadFont()
		// addLabel(templateImage, 100, 100, topText)
		addLabel(templateImage, "top", topText)
		addLabel(templateImage, "bottom", bottomText)

		jpegOptions := jpeg.Options{Quality: 90}
		jpeg.Encode(something, templateImage, jpegOptions)

		imageBuffer := getImageBuffer(templateImage)
		memeLength := len(imageBuffer.Bytes())

		log.Println(reqID, "Generated Meme, length:", memeLength)
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Length", strconv.Itoa(memeLength))

		_, err = w.Write(imageBuffer.Bytes())
		if err != nil {
			log.Println("unable to write image.")
		}
	})
	log.Println("Listening on Port ", strconv.Itoa(PORT))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(PORT), nil))
}

func loadImage(name string) *image.RGBA {
	imageTemplateFile, err := os.Open("./" + name + ".jpg")
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

func addLabel(img *image.RGBA, position string, label string) {
	var x, y int
	switch position {
	case "top":
		x = 100
		y = 100
	case "bottom":
		x = 100
		y = 600
	default:
		panic("add Label function called without valid position")
	}

	context := freetype.NewContext()
	context.SetDst(img)
	size := 12.0 // font size in pixels
	pt := freetype.Pt(x, y+int(context.PointToFixed(size)>>6))

	_, err := context.DrawString(label, pt)
	if err != nil {
		panic("drawing didn't work")
	}
}

func loadFont() *truetype.Font {
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

func getImageBuffer(jpgImage image.Image) *bytes.Buffer {
	imageBuffer := new(bytes.Buffer)
	err := jpeg.Encode(imageBuffer, jpgImage, nil)
	if err != nil {
		panic("unable to encode template")
	}

	return imageBuffer
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
