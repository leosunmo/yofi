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
	Menus     []Menu `yaml:"app"`
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
	Name   string `yaml:"name"`
	Menu   string `yaml:"menu"`
	Script string `yaml:"script"`
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

func (mi *MenuItem) GetScript() (string, bool) {
	if mi.Script != "" {
		return mi.Script, true
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

func (a *App) RunScript(script string) {
	// We could potentiall execute other rofi scripts here
	// args := []string{"-modi", script + ":" + script, "-show", script}
	// //args = append(args, []string{"-format", "p"}...)
	// cmd := exec.Command("rofi", args...)
	cmd := exec.Command(script)
	err := cmd.Run()
	if err != nil {
		// non-zero exit, return to previous menu
		fmt.Println("script err: %s", err.Error())
	}
}

func (a *App) Start() string {
	finalOuput := a.Run(&a.Menus[0])
	return finalOuput
}

func (a *App) Run(m *Menu) string {
	args := []string{"-dmenu"}
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
	c.Wait()

	s := strings.TrimSpace(out.String())
	a.selection = s
	fmt.Printf("Selection was \"%s\"\n", s)
	var sm *MenuItem
	var ok bool
	if sm, ok = m.GetMenuItem(s); !ok {
		log.Fatal("non-menu item was returned")
	}
	// Check if the selection takes us to a new menu
	if mi, ok := sm.GetMenuName(); ok {
		if nm, ok := a.GetMenu(mi); ok {
			a.Run(nm)
		}
	}
	if sc, ok := sm.GetScript(); ok {
		// If this fails, the previous selection is returned as "final",
		// we should check the exit status of the script
		a.RunScript(sc)
		return fmt.Sprintf("Executed script \"%s\"", sc)
	}
	return a.selection
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
	a := NewApp("menu.yaml")
	output := a.Start()
	fmt.Printf("Final selection was %s", output)
}
