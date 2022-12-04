package main

import (
	"apiFP/db"
	"apiFP/identification"
	"apiFP/template/extraction"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/image/bmp"
	"image"
	"image/draw"
	"io/ioutil"
	"net/http"
	"os"
)

type TemplateRawData struct {
	TemplateHEX  string `json:"template_hex"`
	TemplateMeta string `json:"template_meta"`
}

type TemplateRequest struct {
	Templates []TemplateRawData
}

func main() {
	db.ConnectDB()
	identifyFingerprintHandler := http.HandlerFunc(identifyFingerprint)
	addTemplatesHandler := http.HandlerFunc(addTemplates)
	http.Handle("/identify", identifyFingerprintHandler)
	http.Handle("/addTemplates", addTemplatesHandler)
	err := http.ListenAndServe(os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

func addTemplates(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var gotRequest TemplateRequest
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		fmt.Fprintf(w, "err %q\n", err, err.Error())
	} else {
		err = json.Unmarshal(body, &gotRequest)
		if err != nil {
			fmt.Println(w, "can't unmarshal: ", err.Error())
		}
	}
	for i := 0; i < len(gotRequest.Templates); i++ {
		readyData, err := hex.DecodeString(gotRequest.Templates[i].TemplateHEX)
		if err != nil {
			panic(err)
		}
		db.Push(readyData, gotRequest.Templates[i].TemplateMeta)
	}
	return
}

func identifyFingerprint(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	file, _, err := request.FormFile("fingerprint")
	if err != nil {
		message, _ := json.MarshalIndent(map[string]string{"error": "Fingerprint not found"}, "", " ")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(message)
		return
	}

	img, _ := bmp.Decode(file)
	gray := image.NewGray(img.Bounds())
	draw.Draw(gray, img.Bounds(), img, image.Point{}, draw.Src)
	template := extraction.ImgToTemplate(gray)

	jsonResp := identification.IdentifyInParallel(template)

	w.Write(jsonResp)
	return
}
