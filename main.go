package main

import (
	"flag"
	"fmt"
	"github.com/deckarep/gosx-notifier"
	"github.com/howeyc/fsnotify"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	dir    = flag.String("dir", "", "ABSOLUTE path to project's root dir")
	tdir   = flag.String("tdir", "", "RELATIVE path to test dir e.g. src/test/groovy/com/foo")
	all    = flag.Bool("all", false, "run all test files? By default, tests are executed only for the file which is edited")
	note   *gosxnotifier.Notification
)

func executeTest(path string) {
	testClass := extractFileName(path)
	fmt.Printf("Running tests for %s ...\n", testClass)
	cmd := exec.Command("./gradlew", "-Dtest.single="+testClass, "test")
	execute(cmd)
}

func executeTests() {
	fmt.Println("Running all tests ...")
	cmd := exec.Command("./gradlew", "test")
	execute(cmd)
}

func extractFileName(path string) string {
	_, filename := filepath.Split(path)
	return strings.Split(filename, ".")[0]
}

func execute(cmd *exec.Cmd) {
	var out []byte
	var err error

	if out, err = cmd.Output(); err != nil {
		note.Push()
	}
	printReport(out)
}

func printReport(out []byte) {
	fmt.Printf("%s \n", string(out))
}

func main() {
	flag.Parse()
	err := os.Chdir(*dir)

	if err != nil {
		log.Fatal(err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)

	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				if ev.IsModify() {
					if *all {
						executeTests()
					} else {
						executeTest(ev.Name)
					}
				}
			case err := <-watcher.Error:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Watch(*tdir)
	if err != nil {
		log.Fatal(err)
	}

	initNotifications()

	<-done
	watcher.Close()
}
func initNotifications() {
	note = gosxnotifier.NewNotification("Click to view report!!")
	note.Title = "Test failure"
	note.Sender = "com.apple.Safari"
	note.Link = "file://" + *dir + "build/reports/tests/test/index.html"
}
