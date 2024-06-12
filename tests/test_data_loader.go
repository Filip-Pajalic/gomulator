package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Define the struct types matching the JSON structure
type RamEntry struct {
	Address uint16
	Value   uint8
}

type State struct {
	A   uint8     `json:"a"`
	B   uint8     `json:"b"`
	C   uint8     `json:"c"`
	D   uint8     `json:"d"`
	E   uint8     `json:"e"`
	F   uint8     `json:"f"`
	H   uint8     `json:"h"`
	L   uint8     `json:"l"`
	PC  uint16    `json:"pc"`
	SP  uint16    `json:"sp"`
	Ram [][2]uint `json:"ram"`
}

type Cycle struct {
	Address uint16 `json:"value"`
	Value   uint8  `json:"value"`
	Type    string `json:"type"`
}

type Data struct {
	Name    string          `json:"name"`
	Initial State           `json:"initial"`
	Final   State           `json:"final"`
	Cycles  [][]interface{} `json:"cycles"`
}

var Testdata []Data

func LoadJsonTestData(dirPath string, fileName string) {
	// JSON data
	// Load JSON data from file
	filePath := dirPath + "/" + fileName // replace with your JSON file path
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening JSON file:", err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	// Unmarshal the JSON data

	err = json.Unmarshal(byteValue, &Testdata)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	// Use the data

}
