package build

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const maxStack = 100

// CommandFunc represents executable command
type CommandFunc func(ctx context.Context, deps DepsFunc) error

// DepsFunc represents function for executing dependencies
type DepsFunc func(deps ...CommandFunc)

// NewExecutor returns new executor
func NewExecutor(commands map[string]CommandFunc) Executor {
	return Executor{commands: commands}
}

type Executor struct {
	commands map[string]CommandFunc
}

func (e Executor) Paths() []string {
	paths := make([]string, 0, len(e.commands))
	for path := range e.commands {
		paths = append(paths, path)
	}
	return paths
}

func (e Executor) Execute(ctx context.Context, paths []string) error {
	executed := map[reflect.Value]bool{}
	stack := map[reflect.Value]bool{}

	var depsFunc DepsFunc
	errReturn := errors.New("return")
	errChan := make(chan error, 1)
	worker := func(queue <-chan CommandFunc, done chan<- struct{}) {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				var err error
				if err2, ok := r.(error); ok {
					if err2 == errReturn {
						return
					}
					err = err2
				} else {
					err = errors.Errorf("command panicked: %v", r)
				}
				errChan <- err
				close(errChan)
			}
		}()
		for {
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				close(errChan)
				return
			case cmd, ok := <-queue:
				if !ok {
					return
				}
				cmdValue := reflect.ValueOf(cmd)
				if executed[cmdValue] {
					continue
				}
				var err error
				switch {
				case stack[cmdValue]:
					err = errors.New("build: dependency cycle detected")
				case len(stack) >= maxStack:
					err = errors.New("build: maximum length of stack reached")
				default:
					stack[cmdValue] = true
					err = cmd(ctx, depsFunc)
					delete(stack, cmdValue)
					executed[cmdValue] = true
				}
				if err != nil {
					errChan <- err
					close(errChan)
					return
				}
			}
		}
	}
	depsFunc = func(deps ...CommandFunc) {
		queue := make(chan CommandFunc)
		done := make(chan struct{})
		go worker(queue, done)
	loop:
		for _, d := range deps {
			select {
			case <-done:
				break loop
			case queue <- d:
			}
		}
		close(queue)
		<-done
		if len(errChan) > 0 {
			panic(errReturn)
		}
	}

	initDeps := make([]CommandFunc, 0, len(paths))
	for _, p := range paths {
		if e.commands[p] == nil {
			return errors.Errorf("build: command %s does not exist", p)
		}
		initDeps = append(initDeps, e.commands[p])
	}
	func() {
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok && err == errReturn {
					return
				}
				panic(r)
			}
		}()
		depsFunc(initDeps...)
	}()
	if len(errChan) > 0 {
		return <-errChan
	}
	return nil
}

const help = `Crust tool is used to build and run all the applications needed for development and testing on Coreum blockchain

Available commands:
- setup Install all the tools required to develop our software
- build	Builds all the required binaries
- znet 	Tool used to spin up development environment running the same components which are used in production.
- lint	Lints source code and checks that git status is clean
- test  Runs unit tests
- tidy 	Executes go mod tidy
`

// Autocomplete serves bash autocomplete functionality.
// Returns true if autocomplete was requested and false otherwise.
func Autocomplete(executor Executor) bool {
	if prefix, ok := autocompletePrefix(os.Args[0], os.Getenv("COMP_LINE"), os.Getenv("COMP_POINT")); ok {
		autocompleteDo(prefix, executor.Paths(), os.Getenv("COMP_TYPE"))
		return true
	}
	return false
}

// Do receives configuration and runs commands
func Do(ctx context.Context, name string, paths []string, executor Executor) error {
	if len(os.Args) == 1 {
		if _, err := fmt.Fprint(os.Stderr, help, name); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}
	return execute(ctx, paths, executor)
}

func execute(ctx context.Context, paths []string, executor Executor) error {
	pathsTrimmed := make([]string, 0, len(paths))
	for _, p := range paths {
		if p[len(p)-1] == '/' {
			p = p[:len(p)-1]
		}
		pathsTrimmed = append(pathsTrimmed, p)
	}
	return executor.Execute(ctx, pathsTrimmed)
}

func autocompletePrefix(exeName string, cLine, cPoint string) (string, bool) {
	if cLine == "" || cPoint == "" {
		return "", false
	}

	cPointInt, err := strconv.ParseInt(cPoint, 10, 64)
	if err != nil {
		panic(err)
	}

	prefix := strings.TrimLeft(cLine[:cPointInt], exeName)
	lastSpace := strings.LastIndex(prefix, " ") + 1
	return prefix[lastSpace:], true
}

func autocompleteDo(prefix string, paths []string, cType string) {
	choices := choicesForPrefix(paths, prefix)
	switch cType {
	case "9":
		startPos := strings.LastIndex(prefix, "/") + 1
		prefix = prefix[:startPos]
		if len(choices) == 1 {
			for choice, children := range choices {
				if children {
					choice += "/"
				} else {
					choice += " "
				}
				fmt.Println(prefix + choice)
			}
		} else if chPrefix := longestPrefix(choices); chPrefix != "" {
			fmt.Println(prefix + chPrefix)
		}
	case "63":
		if len(choices) > 1 {
			for choice, children := range choices {
				if children {
					choice += "/"
				}
				fmt.Println(choice)
			}
		}
	}
}

func choicesForPrefix(paths []string, prefix string) map[string]bool {
	startPos := strings.LastIndex(prefix, "/") + 1
	choices := map[string]bool{}
	for _, path := range paths {
		if strings.HasPrefix(path, prefix) {
			choice := path[startPos:]
			endPos := strings.Index(choice, "/")
			children := false
			if endPos != -1 {
				choice = choice[:endPos]
				children = true
			}
			if _, ok := choices[choice]; !ok || children {
				choices[choice] = children
			}
		}
	}
	return choices
}

func longestPrefix(choices map[string]bool) string {
	if len(choices) == 0 {
		return ""
	}
	prefix := ""
	for i := 0; true; i++ {
		var ch uint8
		for choice := range choices {
			if i >= len(choice) {
				return prefix
			}
			if ch == 0 {
				ch = choice[i]
				continue
			}
			if choice[i] != ch {
				return prefix
			}
		}
		prefix += string(ch)
	}
	return prefix
}
