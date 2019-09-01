package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

type App struct {
	selection string
	Options   AppOptions `yaml:"options"`
	Menus     []Menu     `yaml:"app"`
}

type AppOptions struct {
	Stdout bool `yaml:"stdout"`
}

// Menu represents a Rofi menu
type Menu struct {
	Name       string     `yaml:"name"`
	Message    string     `yaml:"message"`
	Items      []MenuItem `yaml:"items"`
	Prompt     string     `yaml:"prompt"`
	MarkupRows bool       `yaml:"markup-rows"`
	SelectRow  int        `yaml:"select-row"`
}

type MenuItem struct {
	Name    string      `yaml:"name"`
	Menu    string      `yaml:"menu"`
	Command string      `yaml:"command"`
	Args    []string    `yaml:"args"`
	Confirm interface{} `yaml:"confirm"`
}

type Command struct {
	Executable string
	Args       []string
}

func (c Command) HasArgs() bool {
	return len(c.Args) != 0
}

func (a *App) GetMenu(menuName string) (*Menu, bool) {
	for _, m := range a.Menus {
		if m.Name == menuName {
			return &m, true
		}
	}
	return &Menu{}, false
}

func (m *Menu) GetMenuItem(name string) (*MenuItem, bool) {
	for _, mi := range m.Items {
		if mi.Name == name {
			return &mi, true
		}
	}
	return &MenuItem{}, false

}

func (m *Menu) GetMessage() []string {
	if len(m.Message) > 0 {
		return []string{"-mesg", m.Message}
	}
	return []string{}
}

func (m *Menu) GetPrompt() []string {
	if len(m.Prompt) > 0 {
		return []string{"-p", m.Prompt}
	}
	return []string{}
}

func (mi *MenuItem) GetCommand() (Command, bool) {
	if mi.Command != "" {
		args := []string{}
		if len(mi.Args) != 0 {
			args = mi.Args
		}
		return Command{
			Executable: mi.Command,
			Args:       args,
		}, true
	}
	return Command{}, false
}

func (mi *MenuItem) GetConfirm() (string, bool) {
	switch mi.Confirm.(type) {
	case bool:
		return "", true
	case string:
		return mi.Confirm.(string), true
	}

	return "", false
}

func (mi *MenuItem) GetMenuName() (string, bool) {
	if mi.Menu != "" {
		return mi.Menu, true
	}
	return "", false
}

func rofiArgs(m *Menu) []string {
	var args []string
	args = append(args, m.GetMessage()...)
	args = append(args, m.GetPrompt()...)
	args = append(args, os.Args[1:]...)

	return args
}

func displayRofiError(errMsg string) {
	args := []string{"-e", errMsg}
	c := exec.Command("rofi", args...)
	err := c.Run()
	if err != nil {
		log.Fatalf("failed to display Rofi error message, %s", err.Error())
	}
}

func (a *App) ShowConfirmDialog(msg string) bool {
	confirmMsg := "Are you sure?"
	if msg != "" {
		confirmMsg = msg
	}

	confirmMenu := Menu{
		Message: confirmMsg,
		Prompt:  "Confirm",
		Items: []MenuItem{MenuItem{
			Name: "Yes",
		},
			MenuItem{
				Name: "No",
			},
		},
	}
	resp, err := a.Run(&confirmMenu)
	if err != nil {
		log.Fatalf("Confirmation dialog failed with %s", err.Error())
	}
	if resp == "Yes" {
		return true
	}
	return false

}

func (a *App) ExecuteCommand(cmd Command) (string, error) {
	// Look for command in PATH or local directory. Also checks for executable permissions
	path, execErr := exec.LookPath(cmd.Executable)
	if execErr != nil {
		// File isn't in PATH, not executable or should have ./ in front of it
		return "", execErr
	}

	var args []string
	if cmd.HasArgs() {
		args = cmd.Args
	}

	c := exec.Command(path, args...)
	stdOut := &bytes.Buffer{}
	c.Stdout = stdOut
	cmdErr := c.Run()
	if cmdErr != nil {
		// non-zero exit, return the error message (and Stderr if not empty) for display
		if exitErr, ok := cmdErr.(*exec.ExitError); ok {
			return string(exitErr.Stderr), cmdErr
		}
		return "", cmdErr
	}
	// Do we want to do something with the output?
	//log.Printf("%s successful, output: %s", cmd.Executable, stdOut.String())
	return stdOut.String(), nil
}

func (a *App) Start() (string, error) {
	return a.Run(&a.Menus[0])
}

func (a *App) Run(m *Menu) (string, error) {
	args := []string{"-dmenu", ""}
	//args = append(args, []string{"-format", "p"}...)

	args = append(args, rofiArgs(m)...)
	c := exec.Command("rofi", args...)

	out := &bytes.Buffer{}
	c.Stdout = out
	in, err := c.StdinPipe()
	if err != nil {
		log.Fatal(err.Error())
	}

	sErr := c.Start()
	if sErr != nil {
		log.Fatal(sErr.Error())
	}
	for _, v := range m.Items {
		fmt.Fprintf(in, "%s\n", v.Name)

	}
	in.Close()
	rofiErr := c.Wait()
	if rofiErr != nil {
		exitErr, ok := rofiErr.(*exec.ExitError)
		if ok {
			if exitErr.ProcessState.ExitCode() == 1 {
				if string(exitErr.Stderr) == "" {
					// Rofi exited with code 1, no stderr, probably from pressing one of the kb.canel keys
					///rofi.kb-cancel: Escape,Control+g,Control+bracketleft
					return "", nil
				}
				return "", exitErr
			}
		}
		log.Fatal(rofiErr.Error())
	}
	s := strings.TrimSpace(out.String())
	a.selection = s
	var sm *MenuItem
	var ok bool
	if a.selection == "" {
		//Selection was blank, returning nil
		return "", nil
	}
	if sm, ok = m.GetMenuItem(s); !ok {
		// Non-menu item was selected
		// This could be a bug, or the user typed a search
		// that didn't match anything and pressed enter
		if _, err := a.Run(m); err != nil {
			return "", err
		}
	}
	// Check if the selection takes us to a new menu
	if mi, ok := sm.GetMenuName(); ok {
		if nm, ok := a.GetMenu(mi); ok {
			if _, err := a.Run(nm); err != nil {
				return "", err
			}
		}
	}
	if cmd, ok := sm.GetCommand(); ok {
		var confirmed = true
		if confirmMsg, ok := sm.GetConfirm(); ok {
			confirmed = a.ShowConfirmDialog(confirmMsg)
		}
		// If this fails, the previous selection is returned as "final",
		// we should check the exit status of the command or script
		if confirmed {
			stdout, err := a.ExecuteCommand(cmd)
			if err != nil {
				return "", fmt.Errorf("%s failed with %s\nuse \"./\" for local files or provide an absolute path", cmd, err.Error())
			}
			return stdout, nil
		}
		// Did not confirm, return to previous menu
		if _, err := a.Run(m); err != nil {
			return "", err
		}
	}
	return a.selection, nil
}

func NewApp(file string) App {
	a := App{}
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err.Error())
	}
	err = yaml.Unmarshal(yamlFile, &a)
	if err != nil {
		log.Fatal(err.Error())
	}
	return a
}

func main() {
	var a App
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
	a = NewApp(yamlFile)
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
