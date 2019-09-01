package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/leosunmo/yofi/internal/app"
)

func displayRofiError(errMsg string) {
	args := []string{"-e", errMsg}
	c := exec.Command("rofi", args...)
	err := c.Run()
	if err != nil {
		log.Fatalf("failed to display Rofi error message, %s", err.Error())
	}
}

func main() {
	var yamlFile string
	// Check if config was supplied
	if len(os.Args) < 2 {
		// No explicit yaml file provided, check for menu.yaml
		if info, err := os.Stat("menu.yaml"); err == nil {
			// Menu.yaml is present, but is it a file?
			if !info.IsDir() {
				yamlFile = "menu.yaml"
			} else {
				// Menu.yaml is a directory, fail.
				displayRofiError("Please provide YAML configuration file")
				os.Exit(1)
			}
		} else if os.IsNotExist(err) {
			// There is no menu.yaml, exit
			displayRofiError("Please provide YAML configuration file")
			os.Exit(1)
		} else {
			// Something else has gone wrong, permissions perhaps? Exit and display error if we have any.
			displayRofiError("Please provide YAML configuration file")
			if err != nil {
				log.Fatal(err.Error())
			}
			os.Exit(1)
		}
	} else {
		// A config file was provided as argument, use it.
		yamlFile = os.Args[1]
	}
	a := app.NewApp(yamlFile)
	resp, err := a.Start()
	if err != nil {
		displayRofiError(err.Error())
		log.Fatal(err.Error())
		os.Exit(1)
	}
	if a.Options.Stdout && resp != "" {
		fmt.Print(resp)
	}
	//fmt.Printf("Final selection was %s, now we can do something with that if we want", resp)
}
