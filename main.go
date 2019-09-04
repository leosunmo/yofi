package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

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

type writer struct {
	io.Writer
	timeFormat string
}

func (w writer) Write(b []byte) (n int, err error) {
	return w.Writer.Write(append([]byte(time.Now().Format(w.timeFormat)), b...))
}

func main() {
	log := log.New(&writer{os.Stdout, "2006-01-02 15:04:05"}, " [yofi] ", log.Lshortfile)
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
				log.Print("Can't read the provided configuration file")
				displayRofiError("Please provide YAML configuration file")
				os.Exit(1)
			}
		} else if os.IsNotExist(err) {
			// There is no menu.yaml, exit
			log.Print("No menu configuration file provided")
			displayRofiError("Please provide YAML configuration file")
			os.Exit(1)
		} else {
			// Something else has gone wrong, permissions perhaps? Exit and display error if we have any.
			log.Printf("Error while reading configuration file, err: %s", err.Error())
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
		if a.Options.Stdout && resp != "" {
			fmt.Print(resp)
		}
		displayRofiError(err.Error())
		log.Fatal(err.Error())
		os.Exit(1)
	}
	if a.Options.Stdout && resp != "" {
		fmt.Print(resp)
	}
	//fmt.Printf("Final selection was %s, now we can do something with that if we want", resp)
}
