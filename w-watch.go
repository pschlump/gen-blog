package main

import (
    "fmt"
    "./fsnotify"
    "io/ioutil"
    "log"
	"os/exec"
	"strings"
)
    // "github.com/howeyc/fsnotify"

// Desc: get a list of filenames and directorys
func getFilenames(dir string) (filenames, dirs []string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, nil
	}
	for _, fstat := range files {
		if ! strings.HasPrefix(string(fstat.Name()),".") {
			if fstat.IsDir() {
				dirs = append(dirs, fstat.Name())
			} else {
				filenames = append(filenames, fstat.Name())
			}
		}
	}
	return
}

func main() {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }

    done := make(chan bool)

    // Process events
    go func() {
        for {
            select {
            case ev := <-watcher.Event:
                log.Println("event:", ev)
				fmt.Printf ( "Caught an event\n" );
				oh := exec.Command("./bin/gen-blog")
				out, err := oh.Output()
				if err != nil {
					fmt.Printf("%s\n",err.Error())
				} else {
					fmt.Printf("%s\n",out)
				}
            case err := <-watcher.Error:
                log.Println("error:", err)
            }
        }
    }()

    err = watcher.Watch("pages")
    if err != nil {
        log.Fatal(err)
    }

    err = watcher.Watch("tmpl")
    if err != nil {
        log.Fatal(err)
    }

    err = watcher.Watch("res")
    if err != nil {
        log.Fatal(err)
    }

	// xyzzy - Defect: if get change on ./res - then need to add new "dirs"

	_, dirs := getFilenames ( "./res" )
	for _, v := range dirs {
		// fmt.Printf ( "adding watch for [%s]\n", "res/"+v );
		err = watcher.Watch("res/"+v)
		if err != nil {
			log.Fatal(err)
		}
	}

    <-done

    /* ... do stuff ... */
    watcher.Close()
}
