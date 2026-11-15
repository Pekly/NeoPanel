//go:build windows && !node

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Pekly/NeoPanel/internal/desktop"
)

func main() {
	dataDir := flag.String("data-dir", "", "Desktop data directory (default: LOCALAPPDATA/NeoPanel-Desktop)")
	version := flag.Bool("version", false, "Print version and exit")
	flag.Parse()
	if *version {
		fmt.Println(desktop.Title + " " + desktop.Version)
		return
	}
	dir, err := desktop.DataDirectory(*dataDir)
	if err == nil {
		err = desktop.Run(dir)
	}
	if err != nil {
		desktop.ShowError("NeoPanel Desktop could not start.\n\n" + err.Error() + "\n\nYour server data has not been deleted. Check desktop.log in the app data folder for details.")
		os.Exit(1)
	}
}
