package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const APIEndpoint = "https://api.github.com/repos/"

type Release struct {
	TagName     string    `json:"tag_name"`
	CreatedAt   time.Time `json:"created_at"`
	PublishedAt time.Time `json:"published_at"`
}

type Commits struct {
	SHA    string          `json:"sha"`
	NodeID string          `json:"node_id"`
	Url    string          `json:"html_url"`
	Parent []ParentCommits `json:"parents"`
}

// ParentCommits sub-structure of Commits
type ParentCommits struct {
	Sha     string `json:"sha"`
	Url     string `json:"url"`
	HtmlUrl string `json:"html_url"`
}

// Fetch all the release tags available for stack repository
func getReleases(url string) ([]Release, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	var release []Release
	parseError := json.NewDecoder(resp.Body).Decode(&release)
	defer resp.Body.Close()
	return release, parseError

}

// Fetch all the commits to its corresponding tags available for stack repository
func getCommits(url string) (Commits, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	var commits Commits
	parseError := json.NewDecoder(resp.Body).Decode(&commits)

	defer resp.Body.Close()
	return commits, parseError

}

// Method to indent the JSON and view
func printIndentedJSON(ert interface{}) {
	data, err := json.MarshalIndent(ert, "", "    ")
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}
	fmt.Printf("%s\n", data)
}

func main() {
	username := "Iltwats"
	repoName := "template-template"
	releaseURL := fmt.Sprintf(APIEndpoint+"%s/%s/releases", username, repoName)
	releaseData, err := getReleases(releaseURL)
	if err != nil {
		log.Fatalln(err)
	}
	// extract all the tags
	var tags []string
	for _, val := range releaseData {
		tags = append(tags, val.TagName)
	}
	//fmt.Println(tags)
	tag := tags[2]
	commitsUrl := fmt.Sprintf(APIEndpoint+"%s/%s/commits/%s", username, repoName, tag)
	fmt.Println(commitsUrl)
	commitsResp, comErr := getCommits(commitsUrl)
	if comErr != nil {
		panic(comErr)
	}
	printIndentedJSON(commitsResp)

}
