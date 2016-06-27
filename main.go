package main

import (
    "fmt"
    "os"
    "net/http"
    "log"
    "encoding/json"
    "strings"
    "io/ioutil"
    "path"
    b64 "encoding/base64"
    vips "github.com/daddye/vips"
)

type Configuration struct {
    MediaServer string
}

type ImgParams struct {
	Type string
	Id string
	Date string
	Width uint
	Height uint
}

func main() {
    http.HandleFunc("/media/photos/", handleImage)
    http.HandleFunc("/media/frames/", handleImage)
    http.HandleFunc("/thumbnail/", thumbnailHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func getConfig() (Configuration , error) {
    file, _ := os.Open("config.json")
    config := Configuration{}
    err := json.NewDecoder(file).Decode(&config)
    return config, err
}

func thumbnailHandler(w http.ResponseWriter, r *http.Request) {
    file_name := strings.Trim(r.RequestURI, "/")
    image_hash := strings.Replace(strings.Trim(file_name, ".jpg"), "thumbnail/", "", -1)
    sParams, err := b64.StdEncoding.DecodeString(image_hash)
    if err != nil {
	    http.NotFound(w, r)
    	return
    }

	params := ImgParams{}
    err = json.Unmarshal(sParams, &params);
    if err != nil || params.Date == "" {
	    http.NotFound(w, r)
    	return
    }
    
    var sCat string = ""
    switch params.Type {
    	case "images":
    		sCat = "photos";
    		break;
    	case "video":
    	case "audio":
    		sCat = "frames";
    		break;
    }

	file_path := "media/" + sCat + "/" + params.Date + "/" + params.Id + ".jpg"
    if err := getImage(file_path); err != nil {
	    http.NotFound(w, r)
    	return
    }

    options := vips.Options{
    	Width:	int(params.Width),
    	Height: int(params.Height),
    	Crop:	false,
    	Extend:	vips.EXTEND_WHITE,
    	Interpolator: vips.BILINEAR,
    	Gravity:	vips.CENTRE,
    	Quality:	95,
    }

	file, _ := os.Open(file_path)
	inBuf, _ := ioutil.ReadAll(file)
	buf, err := vips.Resize(inBuf, options)
    if err != nil {
	    http.NotFound(w, r)
    	return
    }

	file.Close();

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Size", string(len(buf)))
	w.Write(buf)

}

func handleImage(w http.ResponseWriter, r *http.Request) {
    file_name := strings.Trim(r.RequestURI, "/")
    if err := getImage(file_name); err == nil {
        http.ServeFile(w, r, file_name)
        return
    }
    http.NotFound(w, r)
}

func getImage(file_name string) error {
    if _, err := os.Stat(file_name); err == nil {
        return err
    }
    config, _ := getConfig()
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
