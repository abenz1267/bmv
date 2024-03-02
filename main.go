package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	fi, _ := os.Stdin.Stat()

	if (fi.Mode() & os.ModeCharDevice) == 0 {
		if len(os.Args) > 1 {
			processor := os.Args[1]

			if processor == "-e" {
				withEditor(getFiles())

				return
			}

			_, err := exec.LookPath(processor)
			if err != nil {
				panic(err)
			}

			withProcessor(os.Args[1:])

			return
		}

		fromStdin()
	} else {
		cmd := exec.Command("ls")
		out, err := cmd.CombinedOutput()
		if err != nil {
			panic(string(out))
		}

		withEditor(strings.Fields(string(out)))
	}
}

func getFiles() []string {
	files := []string{}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Err() != nil {
			log.Panic(scanner.Err())
		}

		files = append(files, scanner.Text())
	}

	return files
}

func withEditor(files []string) {
	editor, ok := os.LookupEnv("EDITOR")
	if !ok {
		fmt.Println("env var 'EDITOR' not set.")
		os.Exit(1)
	}

	dest, err := os.CreateTemp("", "*")
	if err != nil {
		panic(err)
	}

	_, err = dest.WriteString(strings.Join(files, "\n"))
	if err != nil {
		panic(err)
	}

	cmd := exec.Command(editor, dest.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	n, err := os.ReadFile(dest.Name())
	if err != nil {
		panic(err)
	}

	toMove := strings.Fields(string(n))

	if len(toMove) != len(files) {
		panic(fmt.Sprintf("expected %d files, got %d", len(files), len(toMove)))
	}

	for k, v := range files {
		if toMove[k] != v {
			move(v, toMove[k])
		}
	}
}

func withProcessor(args []string) {
	files := getFiles()

	cmd := exec.Command(args[0], args[1:]...)

	pipe, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	go func() {
		defer pipe.Close()
		io.WriteString(pipe, strings.Join(files, "\n"))
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(string(out))
	}

	n := strings.Fields(string(out))

	if len(n) != len(files) {
		panic(fmt.Sprintf("expected %d files, got %d", len(files), len(n)))
	}

	for k, v := range files {
		move(v, n[k])
	}
}

func fromStdin() {
	scanner := bufio.NewScanner(os.Stdin)

	files := getFiles()

	for _, v := range files {
		paths := strings.Fields(v)

		if len(paths) < 2 {
			log.Printf("skipping: %s", scanner.Text())
			continue
		}

		if _, err := os.Stat(paths[0]); errors.Is(err, os.ErrNotExist) {
			log.Printf("file to move doesn't exist: %s", paths[0])
			continue
		}

		move(paths[0], paths[1])
	}
}

func move(src, dest string) {
	if _, err := os.Stat(dest); !errors.Is(err, os.ErrNotExist) {
		log.Printf("cant move '%s' to '%s', as it already exists.", src, dest)
		return
	}

	destDir := filepath.Dir(dest)

	err := os.MkdirAll(destDir, 0755)
	if err != nil {
		log.Println(err)
		return
	}

	cmd := exec.Command("mv", src, dest)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(string(out))
	}

	log.Printf("moved '%s' to '%s'", src, dest)
}
