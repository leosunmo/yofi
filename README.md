# Yofi - YAML defined Rofi menus
Yofi is a simple project that allows you to build Rofi menus in a YAML configuration file.

## Usage
By default yofi looks for a `menu.yaml` file in the current directory, otherwise you can specify the YAML file as an argument.
For example, to run the shutdown menu, `yofi shutdown-menu.yaml`.

## YAML config options
Here's all of the options available.
```yaml
---
options:
  stdout: true   # Print the final selection or script output to STDOUT
app:
- name: main   # Menu name
  message: |   # Messages displayed at the top of the Rofi box
    hello there!
    This is multiline!
  prompt: 'Select'   # Customise the prompt
  items:
    - name: "Shutdown"   # Name of the button in Rofi
      command: shutdown-script   # Command to execute
      confirm: Sure you want to shutdown?  # A confirmation box with this text will appear when this menu item is selected
    - name: "Other menu"
      menu: other-menu
    - name: script not in path
      command: ./not-in-path   # If you want to execute a local script, use "./", by default it will search for the command in $PATH
    - name: script doesn't exist
      command: doesntexist   # Graceful exit with Rofi error message at the top of the screen.
    - name: i3
      command: i3  # Finds the executable using $PATH
      args:        # Run "i3" which these arguments
        - --version
    - name: fail-script
      command: ./fail.sh
- name: other-menu
  message: This is another menu, select "no" to go back
  items:
    - name: "Yes" 
      return: "This message will return to stdout"   # Customise returned string, if not set, menu "name" is returned
    - name: "No" 
      menu: main   # return to menu "main"
  prompt: 'Y/N'
```

## Building
Download [Go](https://golang.org/dl/) and follow the [instructions](https://golang.org/doc/install#install)

TL;DR installing 1.12.9 on Ubuntu-like systems:
```
wget https://dl.google.com/go/go1.12.9.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.12.9.linux-amd64.tar.gz

echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.basrc # or ~/.profile

go version
```

Clone this repo and build it.
```
git clone https://github.com/leosunmo/yofi.git
cd yofi

go build .
```

Or just use `go get` to install the binary only
```
go get github.com/leosunmo/yofi

which yofi
# ~/go/bin/yofi

```