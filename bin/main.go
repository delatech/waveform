package main

import (
	"log"
	"os"

	"github.com/delatech/waveform"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("Usage:", os.Args[0], "filename.audio")
	}

	waveform.Generate(os.Args[1], os.Stdout)
}
