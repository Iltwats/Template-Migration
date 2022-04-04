package main

import (
	"fmt"
	"log"
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

func main() {

	username := "Iltwats"
	repoName := "template-template"
	releaseURL := fmt.Sprintf(APIEndpoint+"%s/%s/releases", username, repoName)
	releaseData, err := getReleases(releaseURL)
	fmt.Println("Fetching all the release tags...")
	if err != nil {
		log.Fatalln(err)
	}
	// extract all the tags from commit json response
	var tags []string
	for _, val := range releaseData {
		tags = append(tags, val.TagName)
	}

	tagSelectedByUser := tags[0]
	userRepoConsumedTag := tags[1]  // TODO fetch from API
	isUserRepoStackConsumed := true // TODO fetch from API
	fmt.Printf("Current version of the repository: %s\n", userRepoConsumedTag)
	message := fmt.Sprintf("Upgrading this repository from version %s to version %s", userRepoConsumedTag, tagSelectedByUser)
	fmt.Println(message)
	if isUserRepoStackConsumed {
		commitsUrl := fmt.Sprintf(APIEndpoint+"%s/%s/commits/%s", username, repoName, tagSelectedByUser)
		commitsResp, comErr := getCommits(commitsUrl)
		fmt.Println("Fetching all the commits for each tags...")
		if comErr != nil {
			panic(comErr)
		}
		parents := commitsResp.Parent
		var parentUrls []string
		for _, parent := range parents {
			parentUrls = append(parentUrls, parent.HtmlUrl)
		}

		var patchFileUrls []string
		patchFileUrls = append(patchFileUrls, commitsResp.Url)
		for _, url := range parentUrls {
			patchFileUrls = append(patchFileUrls, url)
		}

		isPatchFileDownloaded := savePatchFile(patchFileUrls, tagSelectedByUser)

		if isPatchFileDownloaded {
			applyPatchFile(tagSelectedByUser, len(patchFileUrls))
		}
	}

}

func applyPatchFile(tag string, indices int) {
	branchName := "patch-apply"
	err := CheckoutBranch(branchName)
	if err != nil {
		panic("Checkout " + err.Error())
	}
	var patchFileNames []string
	for i := 0; i < indices; i++ {
		name := fmt.Sprintf("%s-%d.patch", tag, i)
		patchFileNames = append(patchFileNames, name)
	}
	for _, name := range patchFileNames {
		patchErr := ApplyPatch(name)
		if patchErr != nil {
			panic("Patch " + patchErr.Error())
			return
		}
	}
	fmt.Println("All the patch files applied successfully.")
	isCacheCleared := DeleteCache(patchFileNames)
	if isCacheCleared {
		pushErr := pushTheBranch(branchName)
		if pushErr != nil {
			log.Fatalln("Error while pushing ", pushErr)
		}
		fmt.Println("Successfully pushed the branch to remote")
		raiseAPullRequest()
	}

}
