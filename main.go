package main

import (
	"bufio"
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		paths := strings.Fields(scanner.Text())

		if len(paths) < 2 {
			log.Printf("skipping: %s", scanner.Text())
			continue
		}

		if _, err := os.Stat(paths[0]); errors.Is(err, os.ErrNotExist) {
			log.Printf("file to move doesn't exist: %s", paths[0])
			continue
		}

		if _, err := os.Stat(paths[1]); !errors.Is(err, os.ErrNotExist) {
			log.Printf("cant move '%s' to '%s', as it already exists.", paths[0], paths[1])
			continue
		}

		destDir := filepath.Dir(paths[1])

		err := os.MkdirAll(destDir, 0755)
		if err != nil {
			log.Println(err)
			continue
		}

		cmd := exec.Command("mv", paths[0], paths[1])
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Println(string(out))
		}

		log.Printf("moved '%s' to '%s'", paths[0], paths[1])
	}

	if scanner.Err() != nil {
		log.Panic(scanner.Err())
	}
}
