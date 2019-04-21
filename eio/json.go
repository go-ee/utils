package eio

import (
	"encoding/json"
	"log"
	"os"
)

func CreateFileJSON(items interface{}, jsonFile string) (file *os.File, err error) {
	var plantsJSON []byte
	if plantsJSON, err = json.MarshalIndent(items, "", "    "); err != nil {
		log.Fatal("Cannot marshal JSON", err)
		return
	}

	if file, err = os.Create(jsonFile); err != nil {
		log.Fatal("Cannot create file", err)
		return
	}
	log.Println("file written", jsonFile)
	defer file.Close()
	file.Write(plantsJSON)

	return
}
