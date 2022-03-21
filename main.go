package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type release struct {
	tag_name string
}

func main() {

	username := "Iltwats"
	repoName := "template-template"
	releaseTags := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", username, repoName)
	resp, err := http.Get(releaseTags)
	if err != nil {
		log.Fatalln(err)
	}
	releaseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	jso := string(releaseData)
	res := json.Unmarshal(releaseData, *release)
	//fmt.Println(json)
	var tag []map[string]string
	jsonMap := map[string]string{}
	for i := range json {
		nm := string(json[i])
		if strings.Contains(nm, "tag-name") {
			jsonMap["tag-name"] = nm
		}
		tag = append(tag, jsonMap)
	}
	fmt.Println(tag[0])
	//url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s", username, repoName, tag)
	////fmt.Println(url)
	//response, err := http.Get(url)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//data, err := ioutil.ReadAll(response.Body)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//json := string(data)
	//fmt.Println(json)
}
