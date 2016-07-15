package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/fatih/color"
)

func Contains(elm string, list []string) bool {
	for _, v := range list {
		if v == elm {
			return true
		}
	}
	return false
}

func Filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func MergeMaps(maps ...map[string]string) map[string]string {
	out := make(map[string]string, 0)
	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

func RunCmdf(cmdline string, a ...interface{}) error {
	return RunCmd(fmt.Sprintf(cmdline, a...))
}

func RunCmd(cmdline string) error {
	log.Println(color.YellowString("> %s", cmdline))
	cmd := exec.Command("/bin/bash", "-c", cmdline)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func MapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		if w, ok := b[k]; !ok || v != w {
			return false
		}
	}

	return true
}

type BySortIndex []map[string]string

func (a BySortIndex) Len() int           { return len(a) }
func (a BySortIndex) Less(i, j int) bool { return a[i]["sortIndex"] < a[j]["sortIndex"] }
func (a BySortIndex) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
