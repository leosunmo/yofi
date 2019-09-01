package app

import "os"

// Menu represents a Rofi menu
type Menu struct {
	Name       string     `yaml:"name"`
	Message    string     `yaml:"message"`
	Items      []MenuItem `yaml:"items"`
	Prompt     string     `yaml:"prompt"`
	MarkupRows bool       `yaml:"markup-rows"`
	SelectRow  int        `yaml:"select-row"`
}

// MenuItem is an individual menu selection with all the available commands and actions
type MenuItem struct {
	Name         string      `yaml:"name"`
	Menu         string      `yaml:"menu"`
	Cmd          string      `yaml:"command"`
	Args         []string    `yaml:"args"`
	Confirm      interface{} `yaml:"confirm"`
	ReturnString string      `yaml:"return"`
}

// Command is an executable (local or through PATH) and it's optional arguments
type Command struct {
	Executable string
	Args       []string
}

// MenuItemByName returns a menu item by its name
func (m *Menu) MenuItemByName(name string) (*MenuItem, bool) {
	for _, mi := range m.Items {
		if mi.Name == name {
			return &mi, true
		}
	}
	return &MenuItem{}, false

}

// MessageArg returns the Menu's message with the appropriate Rofi arguments to display it
func (m *Menu) MessageArg() []string {
	if len(m.Message) > 0 {
		return []string{"-mesg", m.Message}
	}
	return []string{}
}

// PromptArg returns the Menu's prompt with the appropriate Rofi arguments to display it
func (m *Menu) PromptArg() []string {
	if len(m.Prompt) > 0 {
		return []string{"-p", m.Prompt}
	}
	return []string{}
}

// SelectedCommand returns a Command struct if one was provided for this menu item
// otherwise returns false
func (mi *MenuItem) SelectedCommand() (Command, bool) {
	if mi.Cmd != "" {
		args := []string{}
		if len(mi.Args) != 0 {
			args = mi.Args
		}
		return Command{
			Executable: mi.Cmd,
			Args:       args,
		}, true
	}
	return Command{}, false
}

// ConfirmDialog returns the confirmation message and true if one was specified.
func (mi *MenuItem) ConfirmDialog() (string, bool) {
	switch mi.Confirm.(type) {
	case bool:
		return "", true
	case string:
		return mi.Confirm.(string), true
	}

	return "", false
}

// SelectedMenu returns the referenced menu by name if one was specified for this menu item
func (mi *MenuItem) SelectedMenu() (string, bool) {
	if mi.Menu != "" {
		return mi.Menu, true
	}
	return "", false
}

// HasArgs returns true if the Command has arguments specified
func (c Command) HasArgs() bool {
	return len(c.Args) != 0
}

func rofiArgs(m *Menu) []string {
	var args []string
	args = append(args, m.MessageArg()...)
	args = append(args, m.PromptArg()...)
	args = append(args, os.Args[1:]...)

	return args
}
