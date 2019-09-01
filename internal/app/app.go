package app

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

// App represents an instance of an application containing one or more menus
type App struct {
	selection string
	Options   Options `yaml:"options"`
	Menus     []Menu  `yaml:"app"`
}

// Options are application level configuration
type Options struct {
	Stdout bool `yaml:"stdout"`
}

// NewApp returns an app built from the YAML file provided
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

// MenuByName returns a menu by its name
func (a *App) MenuByName(menuName string) (*Menu, bool) {
	for _, m := range a.Menus {
		if m.Name == menuName {
			return &m, true
		}
	}
	return &Menu{}, false
}

// Start starts the application at the first menu (App.Menus[0])
func (a *App) Start() (string, error) {
	return a.Run(&a.Menus[0])
}

// Run executes the provided Menu
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
	if sm, ok = m.MenuItemByName(s); !ok {
		// Non-menu item was selected
		// This could be a bug, or the user typed a search
		// that didn't match anything and pressed enter
		if _, err := a.Run(m); err != nil {
			return "", err
		}
	}
	// Check if the selection takes us to a new menu
	if mi, ok := sm.SelectedMenu(); ok {
		if nm, ok := a.MenuByName(mi); ok {
			if _, err := a.Run(nm); err != nil {
				return "", err
			}
		}
	}
	if cmd, ok := sm.SelectedCommand(); ok {
		var confirmed = true
		if confirmMsg, ok := sm.ConfirmDialog(); ok {
			confirmed = a.ShowConfirmation(confirmMsg)
		}
		// If this fails, the previous selection is returned as "final",
		// we should check the exit status of the command or script
		if confirmed {
			stdout, err := a.ExecuteCommand(cmd)
			if err != nil {
				return "", fmt.Errorf("%s: %s", cmd, err.Error())
			}
			return stdout, nil
		}
		// Did not confirm, return to previous menu
		if _, err := a.Run(m); err != nil {
			return "", err
		}
	}
	if sm.ReturnString != "" {
		a.selection = sm.ReturnString
		return sm.ReturnString, nil
	}
	return a.selection, nil
}

// ShowConfirmation displays a confirmation dialog with the provided message or a default one
// returns true or false depending on user choice of "Yes" or "No"
func (a *App) ShowConfirmation(msg string) bool {
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

// ExecuteCommand takes a Command and executes it. Returns stdout and error if there was any
func (a *App) ExecuteCommand(cmd Command) (string, error) {
	// Look for command in PATH or local directory. Also checks for executable permissions
	path, execErr := exec.LookPath(cmd.Executable)
	if execErr != nil {
		// File isn't in PATH, not executable or should have ./ in front of it
		return "", fmt.Errorf("%s\nuse \"./\" for local files or provide an absolute path", execErr)
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
