// neomind — удобный Go-CLI поверх Rust-ядра.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/narkomaniya/neomind/internal/core"
)

const cliVersion = "0.1.0"

var noColor = os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb"

func paint(code, s string) string {
	if noColor {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func green(s string) string  { return paint("32", s) }
func yellow(s string) string { return paint("33", s) }
func red(s string) string    { return paint("31", s) }
func dim(s string) string    { return paint("90", s) }
func bold(s string) string   { return paint("1", s) }

func usage() {
	fmt.Println(bold("NeoMind") + " — быстрые заметки с ядром на Rust и CLI на Go")
	fmt.Println()
	fmt.Println(bold("Использование:"))
	fmt.Println("  neomind                     show this help")
	fmt.Println("  neomind status [--json]     состояние ядра и хранилища")
	fmt.Println("  neomind note add <текст>    добавить заметку")
	fmt.Println("  neomind note list [--json]  показать заметки")
	fmt.Println("  neomind note clear [--yes]  очистить хранилище")
	fmt.Println("  neomind doctor              проверить установку")
	fmt.Println("  neomind version             версия CLI и ядра")
	fmt.Println()
	fmt.Println(dim("Переменные: NEOMIND_CORE — путь к neomind-core, NO_COLOR — без цвета."))
}

func die(err error) {
	fmt.Fprintln(os.Stderr, red("error:")+" "+err.Error())
	os.Exit(1)
}

func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

func plain(args []string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		if a != "--json" && a != "--yes" {
			out = append(out, a)
		}
	}
	return out
}

func cmdStatus(args []string) int {
	r, err := core.Call("status")
	if err != nil {
		die(err)
	}
	if hasFlag(args, "--json") {
		b, _ := json.MarshalIndent(r, "", "  ")
		fmt.Println(string(b))
		return 0
	}
	fmt.Println(bold("NeoMind") + " " + dim("v"+cliVersion))
	fmt.Printf("  core:       %s\n", green(r.Core+" v"+r.Version))
	fmt.Printf("  cli:        %s\n", green("go v"+cliVersion))
	fmt.Printf("  notes:      %s\n", fmt.Sprint(r.Count))
	fmt.Printf("  data dir:   %s\n", dim(r.DataDir))
	return 0
}

func cmdNote(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "note needs a subcommand: add | list | clear")
		return 2
	}
	switch args[0] {
	case "add":
		text := strings.Join(plain(args[1:]), " ")
		text = strings.TrimSpace(text)
		if text == "" {
			fmt.Fprintln(os.Stderr, "usage: neomind note add <текст>")
			return 2
		}
		r, err := core.Call("note_add", text)
		if err != nil {
			die(err)
		}
		fmt.Println(green("saved") + dim(fmt.Sprintf(" · %d notes total", r.Count)))
		return 0

	case "list":
		r, err := core.Call("note_list")
		if err != nil {
			die(err)
		}
		if hasFlag(args, "--json") {
			b, _ := json.MarshalIndent(r.Notes, "", "  ")
			fmt.Println(string(b))
			return 0
		}
		if len(r.Notes) == 0 {
			fmt.Println(dim("no notes yet — try: neomind note add «первая мысль»"))
			return 0
		}
		for i, n := range r.Notes {
			stamp := time.Unix(n.TS, 0).Format("2006-01-02 15:04")
			fmt.Printf("%s  %s\n", dim(fmt.Sprintf("%3d. %s", i+1, stamp)), n.Text)
		}
		return 0

	case "clear":
		if !hasFlag(args, "--yes") {
			fmt.Print("Remove all notes? [y/N] ")
			ans, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			ans = strings.ToLower(strings.TrimSpace(ans))
			if ans != "y" && ans != "yes" {
				fmt.Println("cancelled")
				return 0
			}
		}
		if _, err := core.Call("note_clear"); err != nil {
			die(err)
		}
		fmt.Println(green("notes cleared"))
		return 0
	}
	fmt.Fprintln(os.Stderr, "unknown note subcommand:", args[0])
	return 2
}

func cmdDoctor() int {
	fmt.Println(bold("NeoMind doctor"))
	code := 0
	p, err := core.Path()
	if err != nil {
		fmt.Println("  " + red("FAIL") + " core binary: " + err.Error())
		return 1
	}
	fmt.Println("  " + green("OK  ") + " core binary: " + dim(p))
	r, err := core.Call("status")
	if err != nil {
		fmt.Println("  " + red("FAIL") + " core answer: " + err.Error())
		return 1
	}
	fmt.Println("  " + green("OK  ") + " core answer: rust v" + r.Version)
	if r.Core != "rust" || r.Version == "" {
		fmt.Println("  " + yellow("WARN") + " unexpected core identity")
		code = 1
	}
	fmt.Println("  " + green("OK  ") + " storage: " + fmt.Sprint(r.Count) + " notes at " + dim(r.DataDir))
	return code
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		return
	}
	switch args[0] {
	case "help", "-h", "--help":
		usage()
	case "version", "-v", "--version":
		r, err := core.Call("version")
		if err != nil {
			fmt.Println("neomind cli v" + cliVersion)
			fmt.Println(dim("core: ") + yellow("not available") + " (" + err.Error() + ")")
			return
		}
		fmt.Println("neomind cli v" + cliVersion)
		fmt.Println("neomind core v" + r.Version + " (rust)")
	case "status":
		os.Exit(cmdStatus(args[1:]))
	case "note":
		os.Exit(cmdNote(args[1:]))
	case "doctor":
		os.Exit(cmdDoctor())
	default:
		fmt.Fprintln(os.Stderr, red("unknown command:")+" "+args[0])
		fmt.Fprintln(os.Stderr, dim("try: neomind help"))
		os.Exit(2)
	}
}
