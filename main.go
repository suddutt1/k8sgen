package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	fmt.Println("Tool started ....")
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Printf("Usage : k8sgen <network json file >\n")
		return
	}
	fmt.Printf("Starting the application.... \n")
	fmt.Printf("Reading the input .... %v\n", args[0])
	configBytes, err := ioutil.ReadFile(args[0])
	if err != nil {
		fmt.Errorf("Error in reading input json")
		return
	}
	GenerateYamlFiles(configBytes)
	fmt.Println("Tool excution complete")

}
