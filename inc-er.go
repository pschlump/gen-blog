package main

//
// inc-er.go -- Program to include other code using templates and our include fuction
// (C) Philip Schlump, 2013.
//

import (
    "./go-flags"
    "io"
    "io/ioutil"
	"log"
	"os"
    "path"
    "regexp"
	"strings"
	"text/template"
)
    // "github.com/jessevdk/go-flags"

func cleanPath ( fn string ) string {
	// top, _ = os.Getwd()
	// sanitize the "fn" to prevent /etc/passwd as an include! or ../../../../../../../../../etc/passwd
	fn = path.Clean(fn)
	fn = strings.Replace(fn, "../", "", -1)
	fn = path.Clean(fn)
	var re = regexp.MustCompile("^/")
	fn = re.ReplaceAllString(fn, "./")
	return fn
}

func literal( b string ) string {
	return  b
}

func encodeTemplateMarkers ( s string ) string {
	//s = strings.Replace(s, "{{", "{{literal \"&#123;&#123;\"}}", -1)
	//s = strings.Replace(s, "}}", "{{literal \"&#125;&#125;\"}}", -1)

	//s = strings.Replace(s, "{{", "{{.StartM&#125;&#125;", -1)
	//s = strings.Replace(s, "}}", "{{.EndM}}", -1)
	//s = strings.Replace(s, "&#125;&#125;", "}}", -1)
	//return s

	r := strings.NewReplacer("{{", "{{.StartM}}", "}}", "{{.EndM}}")
	return r.Replace(s)
}

func include( fn string ) string {

	fn = cleanPath(fn)

    fi, err := os.Open(fn)
    if err != nil { panic(err) }
    defer func() {
        if err := fi.Close(); err != nil {
            panic(err)
        }
    }()

	var rv []byte
    buf := make([]byte, 8192)
    for {
        n, err := fi.Read(buf) // read a chunk
        if err != nil && err != io.EOF { panic(err) }
        if n == 0 { break }

		rv = append ( rv, buf[:n]... )
	}

	return string(rv);
}

func getRawFile(path string) []byte {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}
	return file
}

var  Verbose	[]bool
var  gArgs		[]string

func ParseCmdLineArgs() {
	var opts struct {
		Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
	}

	args, err := flags.ParseArgs(&opts, os.Args)

	if err != nil {
			panic(err)
			os.Exit(1)
	}

	Verbose		= opts.Verbose
	gArgs		= args
}

func main() {

	funcMap := template.FuncMap{
		"title": strings.Title, 
		"include": include,
		"literal": literal,
		"encodeTemplateMarkers": encodeTemplateMarkers,
	}

	ParseCmdLineArgs() 

	templateText :=  getRawFile( gArgs[1] )				// xyzzy - just 1st arg used!!! - should read all!

	tmpl, err := template.New("includeRunner").Delims("[[[[","]]]]").Funcs(funcMap).Parse( string(templateText) )
	if err != nil {
		log.Fatalf("parsing: %s", err)
	}

	err = tmpl.Execute(os.Stdout, "the go programming language")
	if err != nil {
		log.Fatalf("execution: %s", err)
	}

}

