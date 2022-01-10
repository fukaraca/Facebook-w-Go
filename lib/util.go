package lib

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/mingrammer/commonregex"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
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

//SearchForYoutube function checks the string str for any valid youtube links and returns the first encounter with boolean result
func SearchForYoutube(str string) (string, bool) {
	linkler := commonregex.Links(str)
	if len(linkler) == 0 {
		return "", false
	}

	for _, link := range linkler {
		ok, err := regexp.MatchString(`(youtube)|(youtu\.be)`, link)
		if err != nil {
			log.Println("regexp match string failed:", err)
			return "err", false
		}
		if ok {
			return link, ok
		}
	}

	return "", false
}

//GetYtEmbed function gets embed code from YT oembed Api
func GetYtEmbed(shortlink string) (string, bool) {

	reqLink := fmt.Sprintf("https://www.youtube.com/oembed?url=%s&format=json", shortlink)
	resp, err := http.Get(reqLink)
	if err != nil {
		log.Println("Video link json couldn't be get", err)
		return "", false
	}
	defer resp.Body.Close()
	jsonVideo := YtEmbedJson{}
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Video link json couldn't be read", err)
		return "", false
	}
	if string(out) == "Not Found" {
		return "invalidurl", false
	}

	err = json.Unmarshal(out, &jsonVideo)
	if err != nil {
		log.Println("Video link json unmarshal failed", err)
		return "Video can't be loaded", false
	}
	return jsonVideo.Html, true
}

//ListFilesInDir return a slice of n number content/file from given directory root
func ListFilesInDir(root string, n int) ([]string, error) {
	var files []string
	f, err := os.Open(root)
	if err != nil {
		return files, err
	}
	fileInfo, err := f.Readdir(n)
	f.Close()
	if err != nil {
		return files, err
	}

	for _, file := range fileInfo {
		files = append(files, file.Name())
	}
	return files, nil
}
