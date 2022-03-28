package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

func DeleteCache(names []string) bool {
	for _, name := range names {
		_, err := exec.Command("rm", "-rf", name).Output()
		if err != nil {
			fmt.Println("Error while deleting", err)
			return false
		}
	}
	fmt.Println("Patch files cache removed ")
	return true
}

// Method to indent the JSON and view
func printIndentedJSON(ert interface{}) {
	data, err := json.MarshalIndent(ert, "", "    ")
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}
	fmt.Printf("%s\n", data)
}
