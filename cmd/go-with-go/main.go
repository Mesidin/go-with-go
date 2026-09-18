package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/hajimehoshi/ebiten/v2"

	"go-with-go/ui"
)

func main() {
	logf := openLog()
	if logf != nil {
		defer logf.Close()
		log.SetOutput(io.MultiWriter(os.Stderr, logf))
	}
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("panic: %v\n%s", rec, debug.Stack())
			fmt.Fprintf(os.Stderr, "go-with-go crashed. See go-with-go.log next to the executable.\n")
			os.Exit(1)
		}
	}()

	log.Printf("starting")
	ebiten.SetWindowSize(960, 1040)
	ebiten.SetWindowTitle("Go with Go")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(ui.NewApp()); err != nil {
		log.Fatal(err)
	}
}

func openLog() *os.File {
	dir, err := os.Executable()
	if err != nil {
		dir, err = os.Getwd()
		if err != nil {
			return nil
		}
	} else {
		dir = filepath.Dir(dir)
	}
	f, err := os.OpenFile(filepath.Join(dir, "go-with-go.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil
	}
	return f
}
