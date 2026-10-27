package cli

import (
	"bufio"
	"fmt"
	"golang.org/x/term"
	"os"
	"strings"
	"sync"
)

var (
	stdinMu   sync.Mutex
	stdinFile *os.File
	stdinBuf  *bufio.Reader
)

func stdin() *bufio.Reader {
	stdinMu.Lock()
	defer stdinMu.Unlock()
	if stdinBuf == nil || stdinFile != os.Stdin {
		stdinFile, stdinBuf = os.Stdin, bufio.NewReader(os.Stdin)
	}
	return stdinBuf
}



func secret(what, prompt string, confirm bool) (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		line, err := stdin().ReadString('\n')
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			if err != nil {
				return "", fmt.Errorf("no %s on stdin: %w", what, err)
			}
			return "", fmt.Errorf("the %s is empty", what)
		}
		return line, nil
	}
	read := func(p string) (string, error) {
		fmt.Fprint(os.Stderr, p)
		b, err := term.ReadPassword(fd)
		fmt.Fprintln(os.Stderr)
		return string(b), err
	}
	pass, err := read(prompt)
	if err != nil {
		return "", err
	}
	if pass == "" {
		return "", fmt.Errorf("the %s is empty", what)
	}
	if confirm {
		again, err := read("Again: ")
		if err != nil {
			return "", err
		}
		if again != pass {
			return "", fmt.Errorf("the %ss differ", what)
		}
	}
	return pass, nil
}
