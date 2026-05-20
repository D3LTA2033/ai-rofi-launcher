package main

import (
	"fmt"
	"os"
)

const usage = `ai-rofi-config — configure ai-rofi-launcher

usage:
  ai-rofi-config            launch the interactive TUI
  ai-rofi-config serve      start the Talkroom web UI on 127.0.0.1:8765
  ai-rofi-config export     print shell-evaluable env for the launcher
  ai-rofi-config migrate    migrate legacy shell config → YAML
  ai-rofi-config path       print config file path
  ai-rofi-config show       print current config as YAML
  ai-rofi-config help       show this message
`

func main() {
	cmd := ""
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	switch cmd {
	case "", "tui":
		if err := runTUI(); err != nil {
			fail(err)
		}
	case "serve", "web":
		if err := runServe(0); err != nil {
			fail(err)
		}
	case "export":
		cfg, err := Load()
		if err != nil {
			fail(err)
		}
		fmt.Print(cfg.Export())
	case "migrate":
		cfg, ok := tryMigrate()
		if !ok {
			fmt.Fprintln(os.Stderr, "no legacy config at "+LegacyPath())
			os.Exit(1)
		}
		if err := cfg.Save(); err != nil {
			fail(err)
		}
		fmt.Println("migrated → " + ConfigPath())
	case "path":
		fmt.Println(ConfigPath())
	case "show":
		cfg, err := Load()
		if err != nil {
			fail(err)
		}
		data, _ := marshalYAML(cfg)
		fmt.Print(data)
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Fprintln(os.Stderr, "unknown command: "+cmd)
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "error: "+err.Error())
	os.Exit(1)
}
