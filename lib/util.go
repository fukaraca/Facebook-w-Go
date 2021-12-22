package lib

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/disintegration/imaging"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
)

//Striper is a function for trimming whitespaces till input becomes nil if it's necessary
func Striper(str string) *string {
	str = strings.TrimSpace(str)
	if str == "" {
		return nil
	}
	return &str
}

//RandomString function generates and returns n bytes sized string
func RandomString(n int) string {
	randSuffix := make([]byte, n)
	if _, err := rand.Read(randSuffix); err != nil {
		log.Println("random string generation failed:", err)
	}
	return fmt.Sprintf("%X", randSuffix)
}

const highestRes = float32(640)

//ResizeAndSave function resize and saves the form input file to filepath with the given filename.
func ResizeAndSave(file multipart.File, filepath, filename string) error {
	img, err := imaging.Decode(file)
	if err != nil {
		log.Println("image couldn't be decoded:")
		return err
	}

	//aspect ratio
	srcX := float32(img.Bounds().Size().X)
	srcY := float32(img.Bounds().Size().Y)
	if srcX > srcY {
		srcY = srcY * (highestRes / srcX)
		srcX = highestRes
	} else {
		srcX = srcX * (highestRes / srcY)
		srcY = highestRes
	}

	// Resize srcImage to size = highestRes aspect ratio using the Lanczos filter.
	dstImage8060 := imaging.Resize(img, int(srcX), int(srcY), imaging.Lanczos)
	return imaging.Save(dstImage8060, filepath+filename)
}

//GetYtEmbed function gets embed code from YT oembed Api
func GetYtEmbed(shortlink string) string {

	reqLink := fmt.Sprintf("https://www.youtube.com/oembed?url=%s&format=json", shortlink)
	resp, err := http.Get(reqLink)
	if err != nil {
		log.Println("Video link json couldn't be get")
		return "Video can't be loaded!"
	}
	defer resp.Body.Close()
	jsonVideo := YtEmbedJson{}
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Video link json couldn't be read")
		return "Video can't be load"
	}
	err = json.Unmarshal(out, &jsonVideo)
	if err != nil {
		log.Println("Video link json unmarshal failed")
		return "Video can't be load"
	}
	return jsonVideo.Html
}
