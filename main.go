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
    vips "github.com/daddye/vips"
    "strconv"
)

type Configuration struct {
    MediaServer string
}

func main() {
    http.HandleFunc("/download", handleImage)
    http.HandleFunc("/thumbnail/", thumbnailHandler)
    log.Fatal(http.ListenAndServe(":8888", nil))
}

func getConfig() (Configuration , error) {
    file, _ := os.Open("config.json")
    config := Configuration{}
    err := json.NewDecoder(file).Decode(&config)
    return config, err
}

func thumbnailHandler(w http.ResponseWriter, r *http.Request) {
    request_uri := strings.Trim(r.RequestURI, "/thumbnail/")
    request_parts := strings.Split(request_uri, "?")
	image_parts := strings.Split(request_parts[0], "/")
	
	aPaths := map[string]bool {
		"photos" : true,
		"frames" : true,
	}
	
	if len(image_parts) != 3 || !aPaths[image_parts[0]]  {
		http.NotFound(w, r);
	}

	file_path := "media/" + strings.Join(image_parts, "/");
    if err := getImage(file_path); err != nil {
	    http.NotFound(w, r)
    	return
    }
    
	r.ParseForm()
	width := Int(r.Form.Get("width"))
	height := Int(r.Form.Get("height"))
	quality := Int(r.Form.Get("quality"))

	if (width == 0 && height == 0 && quality == 0) {
		http.NotFound(w, r)
		return
	}

    options := vips.Options{
    	Width:	width,
    	Height: height,
    	Crop:	true,
    	Extend:	vips.EXTEND_WHITE,
    	Interpolator: vips.BILINEAR,
    	Gravity: vips.CENTRE,
    	Quality: quality,
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

func Int(v string) int {
	if v == "" {
		return 0;
	}

	val, err := strconv.ParseInt(v, 0, 0)
	if err != nil {
		panic(err)
	}
	return int(val)
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
