package main

import (
	"encoding/json"
	"fmt"
	vips "github.com/daddye/vips"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

var config Configuration

type Configuration struct {
	MediaServer string
}

func initConfig() {
	file, _ := os.Open("/go/src/app/config.json")
	err := json.NewDecoder(file).Decode(&config)
	if err != nil {
		panic(err)
	}
	file.Close()
}

func main() {
	initConfig()
	http.HandleFunc("/thumbnail/", thumbnailHandler)
	log.Fatal(http.ListenAndServe(":8888", nil))
}

func thumbnailHandler(w http.ResponseWriter, r *http.Request) {
	request_uri := strings.Trim(r.RequestURI, "/thumbnail/")
	request_parts := strings.Split(request_uri, "?")
	image_parts := strings.Split(request_parts[0], "/")

	aPaths := map[string]bool{
		"photos": true,
		"frames": true,
		"pubs":   true,
	}

	if !aPaths[image_parts[0]] {
		http.NotFound(w, r)
		return
	}

	file_path := "media/" + strings.Join(image_parts, "/")
	if err := getImage(file_path); err != nil {
		http.NotFound(w, r)
		return
	}

	r.ParseForm()
	width := Int(r.Form.Get("width"))
	height := Int(r.Form.Get("height"))
	quality := Int(r.Form.Get("quality"))
	crop := Bool(r.Form.Get("crop"))
	enlarge := Bool(r.Form.Get("enlarge"))

	if width == 0 && height == 0 && quality == 0 {
		http.NotFound(w, r)
		return
	}

	if width != 0 && height != 0 {
		crop = true
	}

	options := vips.Options{
		Width:        width,
		Height:       height,
		Crop:         crop,
		Extend:       vips.EXTEND_WHITE,
		Enlarge:      enlarge,
		Interpolator: vips.BILINEAR,
		Gravity:      vips.CENTRE,
		Quality:      quality,
	}

	file, file_err := os.Open(file_path)
	if file_err != nil {
		http.NotFound(w, r)
		return
	}

	inBuf, buff_err := ioutil.ReadAll(file)
	if buff_err != nil {
		http.NotFound(w, r)
		return
	}
	
	buf, err := vips.Resize(inBuf, options)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	file.Close()

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Size", string(len(buf)))
	w.Write(buf)
}

func Int(v string) int {
	if v == "" {
		return 0
	}

	val, err := strconv.ParseInt(v, 0, 0)
	if err != nil {
		log.Fatal(err)
		return 0
	}
	return int(val)
}

func Bool(v string) bool {
	if v == "1" || strings.ToLower(v) == "true" {
		return true
	}
	return false
}

func getImage(file_name string) error {
	if _, err := os.Stat(file_name); err == nil {
		return err
	}

	uri := config.MediaServer + "/" + file_name
	return DownloadToFile(uri, file_name)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func HTTPDownload(uri string) ([]byte, error) {
	fmt.Printf("HTTPDownload From: %s.\n", uri)
	res, err := http.Get(uri)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	d, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ReadFile: Size of download: %d\n", len(d))
	return d, err
}

func WriteFile(dst string, d []byte) error {
	fmt.Printf("WriteFile: Size of download: %d\n", len(d))
	os.MkdirAll(path.Dir(dst), os.ModePerm)
	err := ioutil.WriteFile(dst, d, 0444)
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func DownloadToFile(uri string, dst string) error {
	fmt.Printf("DownloadToFile From: %s.\n", uri)
	d, err := HTTPDownload(uri)
	if err == nil {
		fmt.Printf("downloaded %s.\n", uri)
		if WriteFile(dst, d) == nil {
			fmt.Printf("saved %s as %s\n", uri, dst)
		}
	}
	return err
}
