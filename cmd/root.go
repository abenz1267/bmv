/*
Copyright © 2024 Andrej Benz <hello@benz.dev>
*/
package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

var (
	nonFlagArgs []string
	createDirs  bool
	cleanDirs   bool
	bmvFlags    = []string{"-e", "--editor", "--createdirs", "--cleandirs", "-p", "--processor"}
)

var rootCmd = &cobra.Command{
	Use:     "see examples. see 'mv' help.",
	Example: strings.Join([]string{"normal 'mv' actions: bmv oldfile newfile\n", "bmv specific:", "<2 column output from external [src dest\\n]> | bmv", "ls | bmv -e", "ls | bmv -e=vim", "ls | bmv -p sed 's/old/new/'", "bmv -p sed 's/old/new/' [implicit 'ls']"}, "\n"),

	Short: "Wrapper for mv which allows bulk operations",
	Long:  `Utility wrapper for mv which enables bulk operations. Internally calls mv, all flags for mv are supported. For example usage, read the readme at https://github.com/abenz1267/bmv. For more detailed information on the flags, see help for mv.`,
	Run: func(cc *cobra.Command, args []string) {
		fi, _ := os.Stdin.Stat()
		nonFlagArgs = cc.Flags().Args()

		var err error

		createDirs, err = cc.Flags().GetBool("createdirs")
		if err != nil {
			panic(err)
		}

		cleanDirs, err = cc.Flags().GetBool("cleandirs")
		if err != nil {
			panic(err)
		}

		useProcessor, err := cc.Flags().GetBool("processor")
		if err != nil {
			panic(err)
		}

		if (fi.Mode() & os.ModeCharDevice) == 0 {
			editor, err := cc.Flags().GetString("editor")
			if err != nil {
				panic(err)
			}

			if editor != "" {
				withEditor(getFiles(), editor)
				return
			}

			if len(nonFlagArgs) > 0 && useProcessor {
				_, err = exec.LookPath(nonFlagArgs[0])
				if err != nil {
					panic(err)
				}

				withProcessor(getFiles())
				return
			}

			fromStdin()
			return
		}

		editor, err := cc.Flags().GetString("editor")
		if err != nil {
			panic(err)
		}

		if len(nonFlagArgs) == 0 {
			cmd := exec.Command("ls")
			out, err := cmd.CombinedOutput()
			if err != nil {
				panic(string(out))
			}

			if editor == "" {
				editor = "$EDITOR"
			}

			withEditor(strings.Fields(string(out)), editor)
			return
		}

		if useProcessor && len(nonFlagArgs) > 0 {
			_, err := exec.LookPath(nonFlagArgs[0])
			if errors.Is(err, exec.ErrNotFound) || errors.Is(err, fs.ErrPermission) {
				panic(err)
			}

			cmd := exec.Command("ls")
			out, err := cmd.CombinedOutput()
			if err != nil {
				panic(string(out))
			}

			withProcessor(strings.Fields(string(out)))
			return
		}

		passthrough()
	},
}

func init() {
	flags := rootCmd.Flags()

	// mv flags
	backup := newStringValue("", new(string))
	backupFlag := flags.VarPF(backup, "backup", "b", "make a backup of each existing destination file")
	backupFlag.NoOptDefVal = "$VERSION_CONTROL"

	update := newStringValue("", new(string))
	updateFlag := flags.VarPF(update, "update", "u", "control which existing files are updated. See mv --help for more info.")
	updateFlag.NoOptDefVal = "older"

	flags.Bool("debug", false, "explain how a file is copied.  Implies -v")
	flags.BoolP("force", "f", false, "do not prompt before overwriting")
	flags.BoolP("interactive", "i", false, "prompt before overwrite")
	flags.BoolP("no-clobber", "n", false, "do not overwrite an existing file")
	flags.Bool("no-copy", false, "do not copy if renaming fails")
	flags.Bool("strip-trailing-slashes", false, "remove any trailing slashes from each SOURCE argument")
	flags.StringP("suffix", "S", "", "override the usual backup suffix")
	flags.StringP("target-directory", "t", "", "move all SOURCE arguments into DIRECTORY")
	flags.BoolP("no-target-directory", "T", false, "treat DEST as a normal file")
	flags.BoolP("verbose", "v", false, "explain what is being done")
	flags.BoolP("context", "Z", false, "set SELinux security context of destination file to default type")
	flags.Bool("version", false, "output version information and exit")

	// new bmz

	// mv flags
	editor := newStringValue("", new(string))
	editorFlag := flags.VarPF(editor, "editor", "e", "use editor, default: $EDITOR")
	editorFlag.NoOptDefVal = "$EDITOR"

	flags.BoolP("processor", "p", false, "use processor")
	flags.Bool("createdirs", false, "create missing directories")
	flags.Bool("cleandirs", false, "cleanup directories that became empty in the process")
}

func output(reader io.ReadCloser) error {
	buf := make([]byte, 1024)
	for {
		num, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if num > 0 {
			fmt.Printf("%s", string(buf[:num]))
		}
	}
}

func passthrough() {
	mvPath, ok := os.LookupEnv("BMV_MV")
	if !ok {
		mvPath = "/usr/bin/mv"
	}

	if createDirs {
		create(nonFlagArgs[1])
	}

	flags := slices.Clone(os.Args[1:])

	for _, v := range bmvFlags {
		for n, m := range flags {
			if m == v {
				flags = slices.Delete(flags, n, n+1)
			}
		}
	}

	cmd := exec.Command(mvPath, flags...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()

	if cleanDirs {
		clean(nonFlagArgs[0], nonFlagArgs[1])
	}
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Type() string {
	return "string"
}

func (s *stringValue) String() string { return string(*s) }

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

func withEditor(files []string, editor string) {
	if editor == "$EDITOR" {
		var ok bool

		editor, ok = os.LookupEnv("EDITOR")
		if !ok {
			fmt.Println("$EDITOR not set")
			os.Exit(1)
		}
	}

	if len(files) == 0 {
		fmt.Println("no files to edit")
		return
	}

	for k, v := range files {
		info, err := os.Stat(v)
		if err != nil {
			panic(err)
		}

		if info.IsDir() && !strings.HasSuffix(v, "/") {
			files[k] = v + "/"
		}
	}

	slices.Sort(files)

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

	avoidCircularRenames(files, toMove)

	for k, v := range files {
		if toMove[k] != v {
			move(v, toMove[k])
		}
	}
}

func withProcessor(files []string) {
	if len(files) == 0 {
		fmt.Println("no files to edit")
		return
	}

	cmd := exec.Command(nonFlagArgs[0], nonFlagArgs[1:]...)

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
		log.Println(err.Error())
		panic(string(out))
	}

	n := strings.Fields(string(out))

	if len(n) != len(files) {
		panic(fmt.Sprintf("expected %d files, got %d", len(files), len(n)))
	}

	avoidCircularRenames(files, n)

	for k, v := range files {
		move(v, n[k])
	}
}

func fromStdin() {
	for _, v := range getFiles() {
		paths := strings.Fields(v)

		if len(paths) == 2 {
			move(paths[0], paths[1])
		}
	}
}

func move(src, dest string) {
	if src == dest {
		return
	}

	if createDirs {
		create(dest)
	}

	flags := slices.Clone(os.Args[1:])

	for _, v := range bmvFlags {
		for n, m := range flags {
			if m == v {
				flags = slices.Delete(flags, n, n+1)
			}
		}
	}

	for _, v := range nonFlagArgs {
		for n, m := range flags {
			if v == m {
				flags = slices.Delete(flags, n, n+1)
			}
		}
	}

	flags = append(flags, src, dest)

	mvPath, ok := os.LookupEnv("BMV_MV")
	if !ok {
		mvPath = "/usr/bin/mv"
	}

	cmd := exec.Command(mvPath, flags...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	fmt.Println(cmd.String())
	if err != nil {
		log.Println(err)
	}

	if cleanDirs {
		clean(src, dest)
	}
}

func create(dest string) {
	if strings.HasSuffix(dest, string(filepath.Separator)) {
		dest = strings.TrimSuffix(dest, string(filepath.Separator))
	}

	destDir := filepath.Dir(dest)

	err := os.MkdirAll(destDir, 0755)
	if err != nil {
		log.Println(err)
		return
	}
}

func clean(src, dest string) {
	if strings.HasSuffix(dest, string(filepath.Separator)) {
		return
	}

	a := strings.Split(filepath.Dir(src), "/")
	b := strings.Split(filepath.Dir(dest), "/")

	for _, p := range [][]string{a, b} {
		for k, v := range p {
			if v == "." {
				continue
			}

			path := filepath.Join(p[0 : len(p)-k]...)

			_ = os.Remove(path)
		}
	}
}

func avoidCircularRenames(files, renames []string) {
	for b, v := range renames {
		for k, m := range files {
			if v == m && k != b {
				nn := m + "_tmp_900314"
				files[k] = nn
				move(m, nn)
			}
		}
	}
}
