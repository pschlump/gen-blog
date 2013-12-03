package main

//
// gen-blog.go -- Generate an entire static site.
// (C) Philip Schlump, 2013.
// GPL v2 License
//
// Version:  0.5.0
// Date: Thu Nov  7 06:37:40 MST 2013
// Date: Mon Dec  2 16:16:25 MST 2013
//
// This code is still fairly ugly.  This is what happens when you don't know a language
// and you just jump into it.   It is, however getting better.  Lots of rewrites!
//
// The plan is for erison 0.6.0 to clean up the design (i.e. to have one) and to make this
// a normal Go pgrogam with some test code.
//


import (
	"./blackfriday" 				// "github.com/russross/blackfriday"
	"encoding/json"
	"errors"
	"./execX"
	"fmt"
    "./go-flags"
    "html/template"
    "io"
    "io/ioutil"
    "os"
	"os/exec"
	"path"
	"path/filepath"
    "reflect"
	"regexp"
	"sort"
    "strings"
	"time"
	"bytes"
)
// "github.com/jessevdk/go-flags"

var debug_mode bool = false;

type  UsageData struct {
	Base		string
	Seq			string
	Usage		[]*UsageStats
}

type  UsageStats struct {
	Url			string
	Count		int
}

type  CatTagFavList struct {
	Item		string
	HowMany		int						// len(InCategory) exposed to template
	InItem		[]*Pass1APost			// The posts in category
}


type  TemplateData struct {
	Author		string		// Default Author for all pages
	Url			string		// ??? partial URL for page
	SiteUrl		string		// URL for top level of site
	CDNUrl		string		// URL for CDN for images/JavaScript and static content
	StartM		string		// {{
	EndM		string		// }}
	Title		string		// Article title from (1) File name - Processed, (2) JSON Header in article
	HasPrev		bool		// Prev
	UrlPrev		string
	TitlePrev	string
	HasNext		bool		// Next
	UrlNext		string
	TitleNext	string
	UrlThis		string		// This
	rawCategory string		// raw string for Category in A|B|C format
	rawLabel	string		// raw string for Label in A|B|C format
	rawTag		string		// raw string for Tag in A|B|C format
	rawPage		string
	Pages		[]string
	WordCount	int			// Calculated # of words in article
	PubDate		string		// Date from File Name
	PubDateTime	string		// Date from File Name + HH:MI:SS from .md file
	Copyright	string		// (C) Philip Schlump from cfg.json file
	Name		string
	Description	string
	Pageinate	string
	postDir		string
	IncludePath		string		// Path to include from, {{.Dir}}|./tmpl
	RefUrl		string
	MonUrl		string
	State		string
	IsProd		string
	KeyWords	string
}

var g_IncludePath string		// Global to pass the IncludePath to func include
var g_currentDirForThisItem string




type Pass1Data struct {
	AllTags				[]string				// Roll up of all the Tags/Categories/Labels
	AllTagsData			[]CatTagFavList
	AllCategories		[]string
	AllCategoriesData	[]CatTagFavList
	AllLabels			[]string
	AllLabelsData		[]CatTagFavList
	nPosts				int						// total number of posts
	MostRecent			[]*Pass1APost			// The 20 most recent
	MostRecent1			[]*Pass1APost			// The 1 most recent
	MostVisited			[]*Pass1APost			// The 10 most viewd
	Alphebetical		[]*Pass1APost			// Sorted by title
	StartM				template.HTML
	EndM				template.HTML
	Pos					int						// Thie current item in [Alphebetical]			// xyzzy
	Td					TemplateData			// TemplateData -- data not caught by Td
	TdMap				map[string]interface{}
	Url					string					// URL for top level of site
	Top					string					// path to get to TOP of website (relative)
}

type Pass1APost struct {
	Title			string
	Url				string
	Tags			[]string
	Categories		[]string
	Labels			[]string
	PubDate			string
	Author			string
	WordCount		int						// # of words in this
	Summary			template.HTML
	SummaryText		string
	BodyMostRec		template.HTML			// Only keep body for most recent article!
	IsBodyMostRec	bool
	NVisits			int						// How many visits to this URL
	Td				TemplateData
	TdMap			map[string]string		// TemplateData
	State			string
	KeyWords		string
}


var TMPL_DIR = "./tmpl/"

var  Verbose	[]bool
var  Port		string
var  ConfigFile	string
var  doClean	bool
var  authorMode	bool
var  debug		bool						// No moths allowed in code.
var  gArgs		[]string
var globalConfigData map[string]string

var globalUsageData UsageData

func getNViews ( url string ) int {
	for _, v := range globalUsageData.Usage {
		if url == v.Url {
			return v.Count
		}
	}
	return 0
}

func ParseCmdLineArgs() {
	var opts struct {
		Verbose		[]bool	`short:"v" long:"verbose" description:"Show verbose debug information"`
		Port		string	`short:"p" long:"port" description:"A port number to listen on" default:"8080"`
		File		string	`short:"f" long:"file" description:"A file" default:"cfg.json"`
		Go			bool	`short:"g" long:"go" description:"Ready to go to production" default:"false"`
		Debug		bool	`short:"D" long:"debug" description:"Debug flag" default:"false"`
		Clean		bool	`short:"c" long:"clean" description:"clean up before running" default:"false"`
		Author		bool	`short:"a" long:"author" description:"Author Mode, Develop Articles (include drafts)." default:"false"`
	}

	args, err := flags.ParseArgs(&opts, os.Args)

	if err != nil {
			panic(err)
			os.Exit(1)
	}

	Verbose		= opts.Verbose
	Port		= opts.Port
	ConfigFile  = opts.File
	if opts.Go {
		ConfigFile  = "prod."+opts.File
	}
	doClean		= opts.Clean
	debug		= opts.Debug
	authorMode	= opts.Author
	gArgs		= args
}







var  fileExistsPath string = ""

func fileExists( fn string ) bool {
	// dir, _ := os.Getwd()
	// fmt.Printf ( "Looking for %s/%s in path: %s", fileExistsPath, fn, dir )
	if Exists(fileExistsPath+"/"+fn) {
		// fmt.Printf ( "... Was Found!!!\n" );
		// return "Found("+fn+")"
		return true
	}
	// fmt.Printf ( "... Nope\n" );
	// return ""
	return false
}

// <description>{{.Summary | StripHtml | TruncateWords50 | XML_escape }}</description>
// <content:encoded>{{.Summary | XML_escape }}</content:encoded>

func StripHtml( b string ) string {
	// fmt.Printf("input:[%s]\n",b);
	var re = regexp.MustCompile("<[^>]+>")
	b = re.ReplaceAllString(b, "")
	// fmt.Printf("output:[%s]\n",b);
	return b
}

func TruncateWords50 ( b string ) string {
	words := strings.Fields(b)
	return strings.Join(words[0:Min(50,len(words))]," ")
}

func XML_escape ( b string ) template.HTML {
	return template.HTML("<![CDATA["+b+"]]>")
}

// ------------------------------------------------------------------------------------------------------------------
// Globals for Templates (oooh Ick!)
//		{{g "name"}}  Access a global and return its value from an "interface" of string
//		{{set "name=Value"}} Set a value to constant Value
//		{{ bla | set "name"}} Set a value to Value of pipe
// ------------------------------------------------------------------------------------------------------------------
var global_data	map[string]string
func global_init () {
	global_data = make(map[string]string)
}
func global_g ( b string ) string {
	return global_data[b]
}
func global_set ( args ...string ) string {
	if ( len(args) == 1 ) {
		b := args[0]
		var re = regexp.MustCompile("([a-zA-Z_][a-zA-Z_0-9]*)=(.*)")
		x := re.FindAllStringSubmatch(b, -1)
		if ( len(x) == 0 ) {
			name := x[0][1]
			value := ""
			global_data[name] = value
		} else {
			name := x[0][1]
			value := x[0][2]
			global_data[name] = value
		}
	} else if ( len(args) == 2 ) {
		name := args[0]
		value := args[1]
		global_data[name] = value
	} else {
		name := args[0]
		value := strings.Join ( args[1:], "" )
		global_data[name] = value
	}
	return ""
}



// -----------------------------------------------------------------------------
//
// Function: "eq"
// I didn't write "eq" - I think that it is amazing!
//
// -----------------------------------------------------------------------------
//
// From:  https://groups.google.com/forum/#!topic/golang-nuts/OEdSDgEC7js 
// By:Russ Cox  
//
//   Comment by:
//   Rob 'Commander' Pike 	
//   Date: 4/4/12
//   
//   I would like to point out the loveliness of Russ's eq function. The
//   beauty is concentrated in the range loop, where x==y is an interface
//   equality check that verifies type and value equality while avoiding
//   allocation. It is enabled by the "case string, int, ..." etc. line,
//   which does the type check but leaves x of interface type.
//   By the way, there could be more types on that case; if you borrow this
//   function, be sure to add the types you need.
//   
//   -rob 
// 
// I have found this function to be useful in my templates.
// I install it as "eq".
//
// eq reports whether the first argument is equal to one of the
// succeeding arguments.
//
func eq(args ...interface{}) bool {
        if len(args) == 0 {
                return false
        }
        x := args[0]
        switch x := x.(type) {
        case string, int, int64, byte, float32, float64:
                for _, y := range args[1:] {
                        if x == y {
                                return true
                        }
                }
                return false
        }

        for _, y := range args[1:] {
                if reflect.DeepEqual(x, y) {
                        return true
                }
        }
        return false
}



// -----------------------------------------------------------------------------
//
// Function: "gt"
// I didn't write "gt" - This is from the same thread as "eq".
//
// -----------------------------------------------------------------------------
//
// From:  https://groups.google.com/forum/#!topic/golang-nuts/OEdSDgEC7js 
//
// gt returns true if the arguments are of the same type (with int8 and int64
// as the same type) and the first argument is greater than the second. This
// is only defined on string, intX, uintX and floatX all other types return
// false.
//
func gt(a1, a2 interface{}) bool {
	switch a1.(type) {
	case string:
		switch a2.(type) {
		case string:
			return reflect.ValueOf(a1).String() > reflect.ValueOf(a2).String()
		}
	case int, int8, int16, int32, int64:
		switch a2.(type) {
		case int, int8, int16, int32, int64:
			return reflect.ValueOf(a1).Int() > reflect.ValueOf(a2).Int()
		}
	case uint, uint8, uint16, uint32, uint64:
		switch a2.(type) {
		case uint, uint8, uint16, uint32, uint64:
			return reflect.ValueOf(a1).Uint() > reflect.ValueOf(a2).Uint()
		}
	case float32, float64:
		switch a2.(type) {
		case float32, float64:
			return reflect.ValueOf(a1).Float() > reflect.ValueOf(a2).Float()
		}
	}
	return false
}

func ge(a1, a2 interface{}) bool {
	switch a1.(type) {
	case string:
		switch a2.(type) {
		case string:
			return reflect.ValueOf(a1).String() >= reflect.ValueOf(a2).String()
		}
	case int, int8, int16, int32, int64:
		switch a2.(type) {
		case int, int8, int16, int32, int64:
			return reflect.ValueOf(a1).Int() >= reflect.ValueOf(a2).Int()
		}
	case uint, uint8, uint16, uint32, uint64:
		switch a2.(type) {
		case uint, uint8, uint16, uint32, uint64:
			return reflect.ValueOf(a1).Uint() >= reflect.ValueOf(a2).Uint()
		}
	case float32, float64:
		switch a2.(type) {
		case float32, float64:
			return reflect.ValueOf(a1).Float() >= reflect.ValueOf(a2).Float()
		}
	}
	return false
}

func lt(a1, a2 interface{}) bool {
	switch a1.(type) {
	case string:
		switch a2.(type) {
		case string:
			return reflect.ValueOf(a1).String() < reflect.ValueOf(a2).String()
		}
	case int, int8, int16, int32, int64:
		switch a2.(type) {
		case int, int8, int16, int32, int64:
			return reflect.ValueOf(a1).Int() < reflect.ValueOf(a2).Int()
		}
	case uint, uint8, uint16, uint32, uint64:
		switch a2.(type) {
		case uint, uint8, uint16, uint32, uint64:
			return reflect.ValueOf(a1).Uint() < reflect.ValueOf(a2).Uint()
		}
	case float32, float64:
		switch a2.(type) {
		case float32, float64:
			return reflect.ValueOf(a1).Float() < reflect.ValueOf(a2).Float()
		}
	}
	return false
}

func le(a1, a2 interface{}) bool {
	switch a1.(type) {
	case string:
		switch a2.(type) {
		case string:
			return reflect.ValueOf(a1).String() <= reflect.ValueOf(a2).String()
		}
	case int, int8, int16, int32, int64:
		switch a2.(type) {
		case int, int8, int16, int32, int64:
			return reflect.ValueOf(a1).Int() <= reflect.ValueOf(a2).Int()
		}
	case uint, uint8, uint16, uint32, uint64:
		switch a2.(type) {
		case uint, uint8, uint16, uint32, uint64:
			return reflect.ValueOf(a1).Uint() <= reflect.ValueOf(a2).Uint()
		}
	case float32, float64:
		switch a2.(type) {
		case float32, float64:
			return reflect.ValueOf(a1).Float() <= reflect.ValueOf(a2).Float()
		}
	}
	return false
}


func literal( b string ) template.HTML {
	return  template.HTML(b)
}


func leftStart( b string ) template.HTML {
	return  template.HTML("{{")
}


func rightEnd( b string ) template.HTML {
	return  template.HTML("}}")
}

// TestFile: t_clean_path.go
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

func useSearchPath ( fn string ) string {
	// func IsDir(name string) bool {
	// func Exists(name string) bool {
	var re = regexp.MustCompile("{{.Dir}}")
	t := re.ReplaceAllString(g_IncludePath, g_currentDirForThisItem)
	p := strings.Split ( t, "|" )
	for _, v := range p {
		fullFn := v+"/"+fn
		// fmt.Printf ( "useSearchPath for [%s] checking if [%s] exists\n", v, fullFn )
		if  Exists( fullFn ) && ! IsDir ( fullFn ) {
			return fullFn
		}
	}
// xyzzy - should check for file - if not found then report error
	return fn
}

func include( fn string ) template.HTML {

	// fmt.Printf ( "include Top: fn=%s\n", fn );
	fn = cleanPath(fn)
	// fmt.Printf ( "include After cleanPath: fn=%s\n", fn );
	fn = useSearchPath ( fn )
	// fmt.Printf ( "include After useSearchPath: fn=%s\n", fn );

	// xyzzy103
	//+=----------------------------------------------------------------------------------------------=+
	//| really really need to check to see if the file exists - if not give out a decent error message |
	//| that tells where the file was not found and the full path searched for it.  Return an empty    |
	//| string.                                                                                        |
	//+=----------------------------------------------------------------------------------------------=+

    // open input file
    fi, err := os.Open(fn)
    if err != nil { panic(err) }
    // close fi on exit and check for its returned error
    defer func() {
        if err := fi.Close(); err != nil {
            panic(err)
        }
    }()

	var rv []byte
    buf := make([]byte, 8192)
    for {
        // read a chunk
        n, err := fi.Read(buf)
        if err != nil && err != io.EOF { panic(err) }
        if n == 0 { break }

		rv = append ( rv, buf[:n]... )
	}

	r := strings.NewReplacer("<", "&lt;", ">", "&gt;", "&", "&amp;")
	// fmt.Printf ( "include returns ------%s------\n", r.Replace(string(rv)) )
	return template.HTML(r.Replace(string(rv)))
}

func encodeTemplateMarkers ( s string ) string {
	r := strings.NewReplacer("{{", "{{.StartM}}", "}}", "{{.EndM}}")
	return r.Replace(s)
}

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

func summaryOfFile(s string,n int) string {
	var codeBlock = regexp.MustCompile("^```")
	lines := strings.Split(string(s), "\n")
	for i, v := range lines {
		if  codeBlock.MatchString( v ) {
			lines = lines[0:i]
			break
		}
	}
	if ( len(lines) > n ) {
		lines = lines[0:n]
	}
	return strings.Join(lines,"\n")
}

func nWords ( s string ) int {
	words := strings.Fields(s)
	return len(words)
}

// Exists reports whether the named file or directory exists.
func Exists(name string) bool {
    if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		   }
		}
	return true
}
// Really icky way to check to see if it is a Direcotry - but it works.
func IsDir(name string) bool {

    f, err := os.Open(name)
    if err != nil {
        // fmt.Println(err)
        return false
    }
    defer f.Close()
    fi, err := f.Stat()
    if err != nil {
        // fmt.Println(err)
        return false
    }
    switch mode := fi.Mode(); {
    case mode.IsDir():
        // fmt.Println("directory")
		return true
    case mode.IsRegular():
        // fmt.Println("file")
		return false
    }
	return false
}

func dumpVar ( v interface{} ) {
	// s, err := json.Marshal ( v )
	s, err := json.MarshalIndent ( v, "", "\t" )
	if ( err != nil ) {
		fmt.Printf ( "Error: %s\n", err )
	} else {
		fmt.Printf ( "%s\n", s )
	}
}

func varToJson ( v interface{} ) string {
	s, err := json.MarshalIndent ( v, "", "\t" )
	if ( err != nil ) {
		return fmt.Sprintf ( "{ \"error\": \"%s\" }\n", err )			// xyzzy - may need to string encode/repalce " with \\" in err
	} else {
		return string(s)
	}
}


// Return matching items from inArr that match the regular expression 're'
func filterArray( re string, inArr []string ) (outArr []string) {
	var validID = regexp.MustCompile(re)

	outArr = make([]string, 0, len(inArr))
	for k := range inArr {
		if  validID.MatchString( inArr[k] ) {
			outArr = append(outArr, inArr[k])
		}
	}
	// fmt.Printf ( "output = %v\n", outArr )
	return
}
func filterForFuturePosts( inArr []string ) (outArr []string) {
	// fmt.Printf ( "input = %v\n", inArr )
	outArr = make([]string, 0, len(inArr))
	for k, fn := range inArr {
		fn = getPubDate(fn)
		if  ! inFuture ( fn ) {
			outArr = append(outArr, inArr[k])
		}
	}
	// fmt.Printf ( "output = %v\n", outArr )
	return
}

func getRawFile(path string) []byte {
	// path = strings.Replace(path, "../", "", -1)
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}
	return file
}

func putRawFile(path string, data []byte ) {
	err := ioutil.WriteFile(path,data,0644)
	if err != nil { panic(err) }
}

func mdToHTML ( data_md []byte ) []byte {
	return blackfriday.MarkdownCommon(data_md)
}

func isMdFile( in string ) bool {
	var validID = regexp.MustCompile(`.*\.md$`)
	return  validID.MatchString( in )
}

//	Makefile_res_tmpl = GetTemplate ( "./tmpl/Makefile_res.tmpl" )
func GetTemplate ( fn string ) []byte {
	var path_fn = TMPL_DIR + fn
	return getRawFile ( path_fn )
}

func executeBashCmd ( cmd string ) (out []byte) {
    oh := exec.Command(cmd)
    out, err := oh.Output()

    if err != nil {
        println(err.Error())
        return
    }
	return
}

func executeBashCmdSlice ( arg []string ) (out []byte) {
    oh := execX.CommandS2(arg)
    out, err := oh.Output()

    if err != nil {
        println(err.Error())
        return
    }
	return
}

func cleanUrl(url string) string {
	if strings.HasSuffix(url, "/") { 
		return url
	} else {
		return url+"/"
	}
}

//	fmt.Println(path.Ext("/a/b/c/bar.css"))			-- Need to clean up and use path.*
func mkURL ( fn string ) string {
	var re = regexp.MustCompile("\\.md\\.html$")
	fn = re.ReplaceAllString(fn, "")
	var re1 = regexp.MustCompile("\\.md$")
	fn = re1.ReplaceAllString(fn, "")
	var re2 = regexp.MustCompile("\\.html$")
	fn = re2.ReplaceAllString(fn, "")
	return fn + ".html"
}

func fixMdHtml_to_Html ( fn string ) string {
	var re = regexp.MustCompile("\\.xml\\.html$")
	fn = re.ReplaceAllString(fn, ".xml")
	var re1 = regexp.MustCompile("\\.html\\.html$")
	fn = re1.ReplaceAllString(fn, ".html")
	var re2 = regexp.MustCompile("\\.md\\.html$")
	return re2.ReplaceAllString(fn, ".html")
}

func fixFnToTitle ( fn string ) string {
	fn = rmDateFromTitle(fn)
	var re = regexp.MustCompile("\\.md\\.html$")
	fn = re.ReplaceAllString(fn, "")
	var re1 = regexp.MustCompile("\\.md$")
	fn = re1.ReplaceAllString(fn, "")
	var re2 = regexp.MustCompile("\\.html$")
	fn = re2.ReplaceAllString(fn, "")
	fn = strings.Title ( strings.Replace( fn, "-"," ",-1) )
	return fn
}

func genHtmlOrOtherFile ( baseFn, dfltExtention string ) string {
	var re = regexp.MustCompile("\\.md$")
	x := re.FindAllString(baseFn, -1)
	if ( len(x) > 0 ) {
		return baseFn + dfltExtention;
	}
	return baseFn;
}

// This removes a leading set of numbers (usually the date) from the file name that constitutes the title.
func rmDateFromTitle ( fn string ) string {
	// var re = regexp.MustCompile("[0-9]+[-][0-9]+[-][0-9]+[-]")									// part of the cfg.json file
	var re = regexp.MustCompile(globalConfigData["FileDateRegExp"]+"[-]")
	return re.ReplaceAllString(fn, "")
}
func getPubDate ( fn string ) string {
	// var re = regexp.MustCompile("([0-9]+[-][0-9]+[-][0-9]+)")									// part of the cfg.json file
	var re = regexp.MustCompile("("+globalConfigData["FileDateRegExp"]+")")
	var r = re.FindAllString(fn,1)
	if len(r) > 0 {
		return r[0]
	} else {
		return globalConfigData["InitialPubDate"]
	}
}

func date_XML_format ( d string ) string {
	t, err := time.Parse("2006-01-02",d)
    if err != nil {
		return "-- invalid date --"
	}
	s := t.Format ( "2006-01-02T15:04:05Z07:00"  )
	return s
}
func date_US_Short ( d string ) string {
	t, err := time.Parse("2006-01-02",d)
    if err != nil {
		return "-- invalid date --"
	}
	s := t.Format ( "Mon, 02 Jan 2006" )
	return s
}
func date_US_Long ( d string ) string {
	t, err := time.Parse("2006-01-02",d)
    if err != nil {
		return "-- invalid date --"
	}
	s := t.Format ( "Monday, 02 January 2006" )
	return s
}

// Convert a Date into XML Schema (ISO 8601) format.
// 2008-11-17T13:07:54-08:00
// date_ISO_8601
func date_ISO_8601 ( d string ) string {
	t, err := time.Parse("2006-01-02",d)
    if err != nil {
		return "-- invalid date --"
	}
	s := t.Format ( time.RFC3339 )
	return s
}

// Date to RFC-822 Format
// Convert a Date into the RFC-822 format used for RSS feeds.
// {{ site.time | date_to_rfc822 }}
// Mon, 17 Nov 2008 13:07:54 -0800
// date_RFC_822
func date_RFC_822 ( d string ) string {
	t, err := time.Parse("2006-01-02",d)
    if err != nil {
		return "-- invalid date --"
	}
	s := t.Format ( time.RFC822 )
	return s
}


// Date to String
// Convert a date to short format.
// {{ site.time | date_to_string }}
// 17 Nov 2008
// date_Short
func date_Short ( d string ) string {
	t, err := time.Parse("2006-01-02",d)
    if err != nil {
		return "-- invalid date --"
	}
	s := t.Format ( "02 Jan 2006" )
	return s
}

// Date to Long String
// Format a date to long format.
// {{ site.time | date_to_long_string }}
// 17 November 2008
// date_Long
func date_Long ( d string ) string {
	t, err := time.Parse("2006-01-02",d)
    if err != nil {
		return "-- invalid date --"
	}
	s := t.Format ( "02 January 2006" )
	return s
}

func extractLeadingJSONFromPost ( p []byte ) ( js string , post []byte ) {
	var re = regexp.MustCompile("(?s)^[ \t\f\n]*`?`` JSON([^`]*)```")
	var s = string(p)
	var r = re.FindAllString(s,1)
	// fmt.Printf ( "r = %v\nlen(r) = %d\n", r, len(r) )
	if len(r) > 0 {
		var j = r[0]
		// fmt.Printf ( "j before: [%s]\n", j )
		js = strings.Replace ( strings.Replace ( j, "``` JSON", "", 1 ), "```", "",  1 )
		// fmt.Printf ( "j after: [%s]\n", js )
		// fmt.Printf ( "len(j) = %d\n", len(j) )
		post = p[len(j):]
		// fmt.Printf ( "before return, post = [%s]\n", post )
	} else {
		js = "{}"
		post = p
	}
	// fmt.Printf ( "---------- return -----------\n" );
	return
}


func prefixPostfix (  s1 string, b []byte, s2 string ) []byte {
	var l int
	l = len(s1) + len(b) + len(s2) + 1
	rv := make ( []byte, 0, l )
	rv = append ( rv, ([]byte(s1))... )
	rv = append ( rv, b... )
	rv = append ( rv, ([]byte(s2))... )
	return rv
}

// Compare the YYYY-MM-DD passed in 'd' to the time.Now() and return true if it is YYYY-MM-DD is in the future
func inFuture ( d string ) bool {
	n := time.Now()
	t, err := time.Parse("2006-01-02",d)
	if err != nil {
		return true
	}
	if t.After(n) {
		return true
	} else {
		return false
	}
}





///$ {{define "test/min-max.inc"}}

func Min ( a, b int ) int {
	if a < b {
		return a
	}
	return b
}

func Max ( a, b int ) int {
	if a > b {
		return a
	}
	return b
}

///$ {{end}}

















// -----------------------------------------------------------------------------------------------------------------------------------------
// Sort Related Stuff
// -----------------------------------------------------------------------------------------------------------------------------------------

// -------------------------------------------------------------------------------------------------
// By is the type of a "less" function that defines the ordering of its Planet arguments.
// -------------------------------------------------------------------------------------------------
type By func(p1, p2 *Pass1APost) bool

// -------------------------------------------------------------------------------------------------
// Sort is a method on the function type, By, that sorts the argument slice according to the function.
// -------------------------------------------------------------------------------------------------
func (by By) Sort(pass1APost []*Pass1APost) {
	ps := &itemSorter{
		items:   pass1APost,
		by:      by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ps)
}

// -------------------------------------------------------------------------------------------------
// itemSorter joins a By function and a slice of Planets to be sorted.
// -------------------------------------------------------------------------------------------------
type itemSorter struct {
	items   []*Pass1APost
	by      func(p1, p2 *Pass1APost) bool // Closure used in the Less method.
}

// -------------------------------------------------------------------------------------------------
// Len is part of sort.Interface.
// -------------------------------------------------------------------------------------------------
func (s *itemSorter) Len() int {
	return len(s.items)
}

// -------------------------------------------------------------------------------------------------
// Swap is part of sort.Interface.
// -------------------------------------------------------------------------------------------------
func (s *itemSorter) Swap(i, j int) {
	s.items[i], s.items[j] = s.items[j], s.items[i]
}

// -------------------------------------------------------------------------------------------------
// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
// -------------------------------------------------------------------------------------------------
func (s *itemSorter) Less(i, j int) bool {
	return s.by(s.items[i], s.items[j])
}




// -----------------------------------------------------------------------------------------------------------------------------------------
// Slice related stuff.
// -----------------------------------------------------------------------------------------------------------------------------------------

func inSliceString( sl []string, it string ) bool {
	for _, v := range sl {
		if v == it {
			return true
		}
	}
	return false
}

func merge( existing, newData []string ) []string {
	for _, v := range newData {
		if ! inSliceString ( existing, v ) {
			existing = append ( existing, v )
		}
	}
	return existing
}

// -------------------------------------------------------------------------------------------------
// Return T/F if predicate is true when iterating over some set.
// -------------------------------------------------------------------------------------------------
func InSlice(limit int, predicate func(i int) bool) bool {
    for i := 0; i < limit; i++ {
        if predicate(i) {
            return true
        }
    }
    return false
}

// Return a slice of items matching the specified tag.
func findByTag ( aTag string, allOfThePosts []Pass1APost ) ( []*Pass1APost ) {
	found := make ( []*Pass1APost, 0, len(allOfThePosts) )			// The posts that match	
	for k, v := range allOfThePosts {
		// if aTag in v.Tags {
		if len(v.Tags) > 0  {
			if InSlice ( len(v.Tags), func(i int) bool { return v.Tags[i] == aTag } ) {
				found = append ( found, &allOfThePosts[k] )
			}
		}
	}
	return found
}

// Return a slice of items matching the specified category.
func findByLabel ( aTag string, allOfThePosts []Pass1APost ) ( []*Pass1APost ) {
	found := make ( []*Pass1APost, 0, len(allOfThePosts) )			// The posts that match	
	for k, v := range allOfThePosts {
		if len(v.Labels) > 0  {
			if InSlice ( len(v.Labels), func(i int) bool { return v.Labels[i] == aTag } ) {
				found = append ( found, &allOfThePosts[k] )
			}
		}
	}
	return found
}

// Return a slice of items matching the specified label.
func findByCategory ( aTag string, allOfThePosts []Pass1APost ) ( []*Pass1APost ) {
	found := make ( []*Pass1APost, 0, len(allOfThePosts) )			// The posts that match	
	for k, v := range allOfThePosts {
		if len(v.Categories) > 0  {
			if InSlice ( len(v.Categories), func(i int) bool { return v.Categories[i] == aTag } ) {
				found = append ( found, &allOfThePosts[k] )
			}
		}
	}
	return found
}








// -------------------------------------------------------------------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------------------------------------------------------------------
func main() {

	global_init()

	funcMap := template.FuncMap{
		"title":			strings.Title, // The name "title" is what the function will be called in the template text.
		"date_US_Short":	date_US_Short,
		"date_US_Long":		date_US_Long,
		"date_XML_format":	date_XML_format,
		"date_ISO_8601":	date_ISO_8601,
		"date_RFC_822":		date_RFC_822,
		"date_Short":		date_Short,
		"date_Long":		date_Long,
		"include":			include,
		"cleanPath":		cleanPath,
		"literal":			literal,
		"leftStart":		leftStart,
		"rightEnd":			rightEnd,
		"StripHtml":		StripHtml,
		"TruncateWords50":	TruncateWords50,
		"XML_escape":		XML_escape,
		"fileExists":		fileExists,
		"encodeTemplateMarkers": encodeTemplateMarkers,
		"g":				global_g,
		"set":				global_set,
	//	"eq":				eq,						// Included in 1.2 of GO! Yea!
	//	"gt":				gt,
	//	"ge":				ge,
	//	"lt":				lt,
	//	"le":				le,
	}

	ConfigFile = "./cfg.json"

	// -------------------------------------------------------------------------------------------------------------------
	// Part 1 - get the Arguments from the Command Line
	ParseCmdLineArgs()

	// fmt.Printf ( "ConfgFile=%s\n", ConfigFile );

	arg0 := make([]string, 2, 2)
	arg0[0] = "make"
	arg0[1] = "clean"
	fmt.Printf ( "make clean\n" );
	executeBashCmdSlice ( arg0 )

	// -------------------------------------------------------------------------------------------------------------------
	// Part 2 - read in the global config file
	file, e := ioutil.ReadFile(ConfigFile)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	if len(Verbose) > 0 && Verbose[0] {
		fmt.Printf("%s\n", string(file))
	}

	json.Unmarshal(file, &globalConfigData)
	if len(Verbose) > 0 && Verbose[0] {
		fmt.Printf("Results: %v\n", globalConfigData)
	}
	if globalConfigData["InitialPubDate"] == "" { globalConfigData["InitialPubDate"] = "2012-01-01" }
	if globalConfigData["FileDateFormat"] == "" { globalConfigData["FileDateFormat"] = "2012-01-02" }
	if globalConfigData["FileDateRegExp"] == "" { globalConfigData["FileDateRegExp"] = "[0-9]+[-][0-9]+[-][0-9]+" }

	// -------------------------------------------------------------------------------------------------------------------
	// Part 2 - read in the global page usage data
	usage, e8 := ioutil.ReadFile("./freq.json")
	if e8 != nil {
		fmt.Printf("File error: %v\n", e8)
		os.Exit(1)
	}
	json.Unmarshal(usage, &globalUsageData)


	// -------------------------------------------------------------------------------------------------------------------
	// Part 3 - Process the ./res diretory (make etc.)
	var filenames, dirs  []string
	filenames, dirs  = getFilenames ( "./res" )				// xyzzy - filenames is not really used!
	dirs = filterForFuturePosts ( dirs )
	filenames = filterForFuturePosts ( filenames )

	if len(Verbose) > 0 && Verbose[0] {
		fmt.Printf("files = %v\n", filenames )
		fmt.Printf("dirs = %v\n", dirs )
	}

	// -------------------------------------------------------------------------------------------------------------------
	// Part 3 - A - Process Dirs
// xyzzy - this is the section that we need to change
// xyzzy - Make file could just clean out ./posts and ./tmp - at top, then "ln" file into ./posts
// xyzzy - setup {{include "xyzzy"}} to use a search path of 
// xyzzy -    1) the directory it came from
// xyzzy -    2) the ./tmpl directory  - a common library
// xyzzy -    3) the "include_path" in Td
	var top string
	var fn_md []string
	var fn_all []string
	// var Makefile_res_tmpl  []byte
	// Makefile_res_tmpl = GetTemplate ( "Makefile_res.tmpl" )
	// t, err := template.New("makefile_res").Parse(string(Makefile_res_tmpl))
	// if err != nil { panic(err) }
	// fmt.Printf ( "tmpl = %s\n", Makefile_res_tmpl )
	dirHash := make ( map[string]string, 1000 )				// This is a map of file names to directories where the file was found // Limited to 1000!!!
	titleHash := make ( map[string]string, 1000 )			// This is a map of file names to directories where the file was found // Limited to 1000!!!
	top, _ = os.Getwd()
	for k := range dirs {
		// fmt.Printf("dir[%d]=%v\n", k, dirs[k] )
		os.Chdir ( top + "/res/" + dirs[k] )				// Xyzzy - check for errors - if did not CD then do not work
		fn_all, _  = getFilenames ( "." )
		fn_md = filterArray( `.*\.md$`, fn_all )
		for _, vv := range fn_md {
			base := filepath.Base(vv)
			dirHash[base] = "./res/"+dirs[k]
		}
		// fmt.Printf ( "****** before gen-makefile: fn_md = %v\n", fn_md )
		arg1 := make([]string, 1, len(fn_md)+1)
		if ( debug_mode ) {
			arg1[0] = "../../bin/gen-makefile"
		} else {
			arg1[0] = "gen-makefile"
		}
		arg1 = append(arg1, fn_md...)
		// fmt.Printf ( "before gen-makefile: arg1 = %v\n", arg1 )
		// fmt.Printf ( "%v\n", arg1 );
		executeBashCmdSlice ( arg1 )
		// fmt.Printf ( "make # in ./res/%s\n", dirs[k] );
		executeBashCmd ( "make" )
		os.Chdir ( top )
	}

	// fmt.Printf ( "dirHash = %v\n", dirHash )

	// -------------------------------------------------------------------------------------------------------------------
	// Part 3 - B - Process Files




	// -------------------------------------------------------------------------------------------------------------------
	// part 4 - Do each of the "pages" in order.
		// * sort list of posts
		// + Filter for future dates
		// Build list of posts / categories / tags so that each of these is accessable for templates
		// combine post with correct set of libraries of templates
		// process each template into final .html file
	fn_all, _  = getFilenames ( "./posts" )
	sort.Strings(fn_all)
	// xyzzy fn_all = filterFutureDates ( fn_all )
	// xyzzy fn_all = filterDrafts ( fn_all )

	// -------------------------------------------------------------------------------------------------------------------
	pass1Data := Pass1Data{
		AllTags: make( []string, 0, 25 ),
		AllCategories: make( []string, 0, 25 ),
		AllLabels: make( []string, 0, 25 ),
		nPosts: 0,
		MostRecent: make( []*Pass1APost, 0, 20 ),
		MostRecent1: make( []*Pass1APost, 0, 1 ),
		MostVisited: make( []*Pass1APost, 0, 20 ),
		Alphebetical: make( []*Pass1APost, 0, 2 ),
		StartM: template.HTML("{{"),
		EndM: template.HTML("}}"),
		Url: "xyzzy-url",
		Top: "",
	}

	allOfThePosts := make( []Pass1APost, 0, 100 )


	// -------------------------------------------------------------------------------------------------------------------
	templateData := TemplateData{
		StartM: "{{",
		EndM: "}}",
		IncludePath: "{{.Dir}}|./tmpl",
		IsProd: "no",
	}
// xyzzy - might be appropriate to put in data at this point
	for k, v := range globalConfigData { 										//  pull data from cfg.json
		switch ( k ) {
		case "Author":		templateData.Author = v
		case "Title":		templateData.Title = v
		case "Url":			templateData.Url = cleanUrl(v)
		case "SiteUrl":		templateData.SiteUrl = cleanUrl(v)
		case "CDNUrl":		templateData.CDNUrl = cleanUrl(v)
		case "MonUrl":		templateData.MonUrl = cleanUrl(v)
		case "categories":	templateData.rawCategory = v
		case "labels":		templateData.rawLabel = v
		case "tags":		templateData.rawTag = v

		case "pages":		templateData.rawPage = v
							// fmt.Printf ( "templateData.rawPage = [%v]\n", templateData.rawPage );
							templateData.Pages = strings.Split(v,"|")
							// fmt.Printf ( "templateData.Pages = [%v]\n", templateData.Pages );

		case "Copyright":	templateData.Copyright = v
		case "Name":		templateData.Name = v
		case "Description":	templateData.Description = v
		case "Pageinate":	templateData.Pageinate = v
		case "PubDate":		templateData.PubDate = v

		case "State":		templateData.State = v
		case "IsProd":		templateData.IsProd = v
		case "KeyWords":	templateData.KeyWords = v

		/* Ignored Items */
		case "InitialPubDate":
		case "FileDateFormat":
		case "FileDateRegExp":
		default:
			fmt.Printf ( "Global-Config, File:%s did not use %s:%v. Hope that's ok.\n", ConfigFile, k, v )
		}
	}



	// fmt.Printf ( "MonUrl = %s\n", templateData.MonUrl );

	var fn string

	last_pos := -1;
	var mostRecentFnAll string
	mostRecentFnAll = ""



// --------------------------------------------------------------------------------------------------------------------------------------------------
// Pass 1
// --------------------------------------------------------------------------------------------------------------------------------------------------
	for _, aPage := range templateData.Pages {

		if ! Exists(aPage) {

		} else if Exists(aPage) && ! IsDir(aPage) {

		} else {
			fn_all, _  = getFilenames ( "./"+aPage )
			sort.Strings(fn_all)
			// fmt.Printf("fn_all Pass1 = %v\n",fn_all);
			prev_pos := -2;
			pos := -1
			for f := range fn_all {
				// prev_pos = pos;
				pos++
				aa := Pass1APost{ WordCount: 0, NVisits: 0, }

				fn = "./"+aPage+"/" + fn_all[f]
				b := getRawFile(fn)
				fileExistsPath = dirHash[ filepath.Base(fn_all[f]) ]
				fileExistsPath = strings.Replace( fileExistsPath ,"/res/", "/_site/", 1)
				g_currentDirForThisItem = dirHash[ filepath.Base(fn_all[f]) ]
				templateData.RefUrl = "../" + filepath.Base( g_currentDirForThisItem )
				// fmt.Printf ( "before [%s]\n", b )
				js, post := extractLeadingJSONFromPost ( b )						// Strip off any leading JSON chunk - config
				// fmt.Printf ( "js=[%s]\npost=[%s]\n", js, post )
				sum := summaryOfFile(string(post), 15)								// 15 should be a configurable item
				sum_b := mdToHTML ( []byte(sum) )
				// sum_b = prefixPostfix ( "{{define \"summary\"}}\n", sum_b, "\n{{end}}\n" )

				aa.Summary = template.HTML(sum_b)
				aa.SummaryText = string(sum_b)
				// fmt.Printf ( "pos=%d len=%d\n", pos, len(fn_all) )

				// fmt.Printf ( "post [%s]\n", post );
				/// _, x_post := extractLeadingJSONFromPost ( post )
				// fmt.Printf ( "x_post [%s]\n", x_post );
				/// post_b := mdToHTML ( x_post )

				aa.WordCount = nWords ( string(b) )
				aa.Title = fixFnToTitle(fn_all[f])
				aa.PubDate = getPubDate(fn_all[f])
				aa.Url = mkURL ( fn_all[f] )

				// Allow override of some of the global data with the JSON at the top.
				var jsonData map[string]string
				json.Unmarshal( []byte(js), &jsonData)
				if ( jsonData["State"] == "draft" || jsonData["State"] == "idea" ) && ! authorMode {
				} else {
					for k, v := range jsonData {
						switch ( k ) {
						case "Author":		aa.Author = v
						case "State":		aa.State = v
						case "KeyWords":	aa.KeyWords = v
						case "Title":		aa.Title = v
						case "Url":			aa.Url = cleanUrl(v)
						//case "CDNUrl":		aa.CDNUrl = cleanUrl(v)
						case "categories":	aa.Categories = strings.Split(v,"|")
						case "labels":		aa.Labels = strings.Split(v,"|")
						case "tags":		aa.Tags = strings.Split(v,"|")
						case "PubDate":		aa.PubDate = v
						}
					}
					titleHash[fn_all[f]] = aa.Title

					// Save Labels, Tags, Categories
					pass1Data.AllCategories = merge( pass1Data.AllCategories, aa.Categories )
					pass1Data.AllLabels = merge( pass1Data.AllLabels, aa.Labels )
					pass1Data.AllTags = merge( pass1Data.AllTags, aa.Tags )
					// aa.BodyMostRec = template.HTML( post_b )
					aa.BodyMostRec = template.HTML("-- not most recent --")
					aa.IsBodyMostRec = true;
					allOfThePosts = append ( allOfThePosts, aa )
					// fmt.Printf ( "prev_pos = %d, just before if\n", prev_pos );
					if prev_pos >= 0 && prev_pos < pos {
						//	fmt.Printf ( "prev_pos=%d --- pos=%d\n", prev_pos, pos );
						allOfThePosts[prev_pos].IsBodyMostRec = false;
					}
					pass1Data.Alphebetical = append ( pass1Data.Alphebetical, &allOfThePosts[ pass1Data.nPosts ] )
					pass1Data.nPosts++
					// xyzzy - should save "Title" for pass 2 so can access it then for next/prev title -- Note that Title is pulled from the JSON at this point.
					prev_pos = len( allOfThePosts )-1
					last_pos = len( allOfThePosts )-1
					mostRecentFnAll = fn_all[f]
				} /* not a draft */
			}
		}
	}
	for _, pi := range allOfThePosts {
		pi.NVisits = getNViews ( pi.Url );
	}

	// --------------- Sort by Date Descending -----------------------------------------------------------------------
	pass1Data.MostRecent = make( []*Pass1APost, 0, len(allOfThePosts) )
	//for i := 0; i < len(allOfThePosts); i++ {
	for i, _ := range allOfThePosts {
		pass1Data.MostRecent = append ( pass1Data.MostRecent, &allOfThePosts[i] )
	}
	byPostDate := func(p1, p2 *Pass1APost) bool {
		if p1.PubDate == p2.PubDate {
			return p1.Title < p2.Title
		}
		return p1.PubDate > p2.PubDate
	}
	By(byPostDate).Sort(pass1Data.MostRecent)

	pass1Data.MostRecent1 = append ( pass1Data.MostRecent1, pass1Data.MostRecent[0] )
	// fmt.Printf ( "len(pass1Data...)=%d\n", len(pass1Data.MostRecent) )

	// fmt.Printf ( "pass1Data.1382 =\n" );
	// dumpVar ( pass1Data )



	// --------------- Sort by Number of Visits to Page  -----------------------------------------------------------------------
	byVisits := func(p1, p2 *Pass1APost) bool {
		return p1.NVisits > p2.NVisits
	}
	pass1Data.MostVisited = make( []*Pass1APost, 0, len(allOfThePosts) )
	for i, _ := range allOfThePosts {
		pass1Data.MostVisited = append ( pass1Data.MostVisited, &allOfThePosts[i] )
	}
	By(byVisits).Sort(pass1Data.MostVisited)


	// --------------- Sort by Title -----------------------------------------------------------------------
	byTitle := func(p1, p2 *Pass1APost) bool {
		return p1.Title < p2.Title
	}
	pass1Data.Alphebetical = make( []*Pass1APost, 0, len(allOfThePosts) )
	for i, _ := range allOfThePosts {
		pass1Data.Alphebetical = append ( pass1Data.Alphebetical, &allOfThePosts[i] )
	}
	By(byTitle).Sort(pass1Data.Alphebetical)

	// fmt.Printf ( "pass1Data.1422 =\n" );
	// dumpVar ( pass1Data )

	// xyzzy - should sort AllTags, All* before doing this

	for _, tg := range pass1Data.AllTags {
		t := CatTagFavList{
			Item: tg,
			InItem: findByTag ( tg, allOfThePosts ),
		}
		t.HowMany = len(t.InItem)
		pass1Data.AllTagsData = append ( pass1Data.AllTagsData, t )
	}
	for _, tg := range pass1Data.AllLabels {
		t := CatTagFavList{
			Item: tg,
			InItem: findByLabel ( tg, allOfThePosts ),
		}
		t.HowMany = len(t.InItem)
		pass1Data.AllLabelsData = append ( pass1Data.AllLabelsData, t )
	}
	for _, tg := range pass1Data.AllCategories {
		t := CatTagFavList{
			Item: tg,
			InItem: findByCategory ( tg, allOfThePosts ),
		}
		t.HowMany = len(t.InItem)
		pass1Data.AllCategoriesData = append ( pass1Data.AllCategoriesData, t )
	}

// --------------------------------------------------------------------------------------------------------------------------------------------------
// Pass 2
// --------------------------------------------------------------------------------------------------------------------------------------------------
	for _, aPage := range templateData.Pages {

		if ! Exists(aPage) {
			fmt.Printf ( "Warning: the page %s did not exist.  This is specified in %s with the 'pages'\n", aPage, ConfigFile );
		} else if Exists(aPage) && ! IsDir(aPage) {

			fmt.Printf ( "Top[page]: %s\n", aPage );
			pass1Data.Top = ""

			// No Next Prev -------------------------------------------------------------------------------------------------------
			templateData.HasPrev = false
			templateData.UrlPrev = ""
			templateData.TitlePrev = ""
			templateData.HasNext = false
			templateData.UrlNext = ""
			templateData.TitleNext = ""

			fn = aPage
			fileExistsPath = "."
			templateData.postDir = "."
			b := getRawFile(fn)
			// fmt.Printf ( "before [%s]\n", b )
			js, post := extractLeadingJSONFromPost ( b )				// Strip off any leading JSON chunk - config
			// fmt.Printf ( "js=[%s]\npost=[%s]\n", js, post )
			if isMdFile ( fn ) {
				b = mdToHTML ( post )
			} else {
				b = post
			}
			b = prefixPostfix ( "{{define \"body\"}}", b, "{{end}}" )
			fn = fixMdHtml_to_Html ( "./tmp/" + filepath.Base(aPage) + ".html" )
			putRawFile ( fn, b )


			// # of words in the article - kind of a weird count but...
			// fmt.Printf("json = [%s]\n", js )
			templateData.WordCount = nWords ( string(b) )

			// Do the Title default - based on file name.
			// templateData.Title = strings.Title ( strings.Replace(strings.Replace( rmDateFromTitle(aPage) ,".md","",-1),"-"," ",-1) )		
			templateData.Title = fixFnToTitle(aPage)

			// Allow override of some of the global data with the JSON at the top.
			var jsonData map[string]string
			json.Unmarshal( []byte(js), &jsonData)

			if ( jsonData["State"] == "draft" || jsonData["State"] == "idea" ) && ! authorMode {
			} else {
				for k, v := range jsonData {
					switch ( k ) {
					case "Author":		templateData.Author = v
					case "State":		templateData.State = v
					case "KeyWords":	templateData.KeyWords = v
					case "Title":		templateData.Title = v
					case "Url":			templateData.Url = cleanUrl(v)
					case "CDNUrl":		templateData.CDNUrl = cleanUrl(v)
					case "categories":	templateData.rawCategory = v
					case "labels":		templateData.rawLabel = v
					case "tags":		templateData.rawTag = v
					case "PubDate":		templateData.PubDate = v
					case "Copyright":	templateData.Copyright = v
					case "Name":		templateData.Name = v
					case "Description":	templateData.Description = v
					case "Pageinate":	templateData.Pageinate = v
					case "IncludePath": templateData.IncludePath = v
										g_IncludePath = templateData.IncludePath
					default:
						fmt.Printf ( "Per-File-Config, File:%s did not use %s:%v.  Hope that's ok.\n", aPage, k, v )
					}
				}
//xyzzy102
				tfn := genHtmlOrOtherFile ( filepath.Base(aPage), ".html" )
				templateData.UrlThis = fixMdHtml_to_Html( templateData.Url + tfn )

				t, err := template.New("whatever").Funcs(funcMap).ParseFiles ( "./tmpl/common.tmpl", fn )
				if err != nil { panic(err) }

				fo, err := os.Create( fixMdHtml_to_Html("_site/"+ tfn))				// Xyzzy - make this into a func. - the open thing.
				if err != nil { panic(err) }
				defer func() {
					if err := fo.Close(); err != nil {
						panic(err)
					}
				}()

				pass1Data.Td = templateData
				pass1Data.Url = fixMdHtml_to_Html( filepath.Base(aPage)+".html" )
				// fmt.Printf ( "pass1Data 1540 %s =\n", aPage );
				// dumpVar ( pass1Data )
				err = t.ExecuteTemplate(fo, filepath.Base(aPage), pass1Data)
				if err != nil { panic(err) }

				g_IncludePath = "{{.Dir}}|./tmpl"
				templateData.IncludePath = "{{.Dir}}|./tmpl"
			}

		} else {

			fmt.Printf ( "Top[dir]:  %s\n", aPage );
			pass1Data.Top = "../"
			fn_all, _  = getFilenames ( "./"+aPage )
			sort.Strings(fn_all)

			pos := -1
			for f := range fn_all {
				pos++

				fmt.Printf ( "------------------- File: %s \n", fn_all[f] );

				// Generate Next Prev -------------------------------------------------------------------------------------------------------
				templateData.HasPrev = false
				templateData.UrlPrev = ""
				templateData.TitlePrev = ""
				templateData.HasNext = false
				templateData.UrlNext = ""
				templateData.TitleNext = ""
				if len(fn_all) > 1 {
					if f < len(fn_all)-1 {
						templateData.HasNext = true
						templateData.UrlNext = mkURL ( fn_all[f+1] )
						templateData.TitleNext = titleHash[fn_all[f+1]]
					}
					if f > 0 {
						templateData.HasPrev = true
						templateData.UrlPrev = mkURL ( fn_all[f-1] )
						templateData.TitlePrev = titleHash[fn_all[f-1]]
					}
				}

				// fmt.Printf ( "NextURL [%s] PrevUrl [%s]\n", templateData.UrlNext , templateData.UrlPrev )

				fn = "./"+aPage+"/" + fn_all[f]
				b := getRawFile(fn)

				// fileExistsPath = filepath.Dir(fn)
				fileExistsPath = dirHash[ filepath.Base(fn_all[f]) ]
				fileExistsPath = strings.Replace( fileExistsPath ,"/res/", "/_site/", 1)
				g_currentDirForThisItem = dirHash[ filepath.Base(fn_all[f]) ]
				templateData.RefUrl = "../" + filepath.Base( g_currentDirForThisItem )
				// fmt.Printf ( "setting .RefUrl to [%s]\n", templateData.RefUrl )
				// fmt.Printf ( "fileExistsPath [%s] for [%s]\n", fileExistsPath, filepath.Base(fn_all[f]) )
				// fmt.Printf ( "before [%s]\n", b )
				js, post := extractLeadingJSONFromPost ( b )						// Strip off any leading JSON chunk - config
				// fmt.Printf ( "js=[%s]\npost=[%s]\n", js, post )
				sum := summaryOfFile(string(post), 15)								// 15 should be a configurable item
				sum_b := mdToHTML ( []byte(sum) )
				sum_b = prefixPostfix ( "{{define \"summary\"}}\n", sum_b, "\n{{end}}\n" )

				b = mdToHTML ( post )
				b = prefixPostfix ( "{{define \"body\"}}\n", b, "\n{{end}}\n" )

				b = append ( b, sum_b... )

				fn = fixMdHtml_to_Html ( "./tmp/" + fn_all[f] + ".html" )
				putRawFile ( fn, b )


				// # of words in the article - kind of a weird count but...
				// fmt.Printf("json = [%s]\n", js )
				templateData.WordCount = nWords ( string(b) )

				// Do the Title default - based on file name.
				// templateData.Title = strings.Title ( strings.Replace(strings.Replace( rmDateFromTitle(fn_all[f]) ,".md","",-1),"-"," ",-1) )		
				templateData.Title = fixFnToTitle(fn_all[f])

				templateData.PubDate = getPubDate(fn_all[f])
				templateData.postDir = fileExistsPath

				// Allow override of some of the global data with the JSON at the top.
				var jsonData map[string]string
				json.Unmarshal( []byte(js), &jsonData)

				if ( jsonData["State"] == "draft" || jsonData["State"] == "idea" ) && ! authorMode {
				} else {

					for k, v := range jsonData {
						switch ( k ) {
						case "Author":		templateData.Author = v
						case "State":		templateData.State = v
						case "KeyWords":	templateData.KeyWords = v
						case "Title":		templateData.Title = v
						case "Url":			templateData.Url = cleanUrl(v)
						case "CDNUrl":		templateData.CDNUrl = cleanUrl(v)
						case "categories":	templateData.rawCategory = v
						case "labels":		templateData.rawLabel = v
						case "tags":		templateData.rawTag = v
						case "PubDate":		templateData.PubDate = v
						case "Copyright":	templateData.Copyright = v
						case "Name":		templateData.Name = v
						case "Description":	templateData.Description = v
						case "Pageinate":	templateData.Pageinate = v
						case "IncludePath": templateData.IncludePath = v
											g_IncludePath = templateData.IncludePath
						default:
							fmt.Printf ( "Per-File-Config, File:%s did not use %s:%v.  Hope that's ok.\n", fn_all[f], k, v )
						}
					}
					templateData.UrlThis = fixMdHtml_to_Html( templateData.Url + aPage + "/" + fn_all[f] + ".html" )

					// check for existinence of ./tmpl/common_{{aPage}}.tmpl
					if ( ! Exists ( "./tmpl/common_"+aPage+".tmpl" ) ) {
						err := errors.New("Missing file: ./tmpl/common_"+aPage+".tmpl" );
						panic(err)
					}

					t, err := template.New("whatever").Funcs(funcMap).ParseFiles ( "./tmpl/common.tmpl", "./tmpl/common_"+aPage+".tmpl", fn )
					if err != nil { panic(err) }

					// make directory for _site/{{aPage}} if is not there already
					if ( ! Exists ( "./_site/"+aPage ) ) {
						os.MkdirAll ( "./_site/"+aPage, 0755 );
					}

					fo, err := os.Create( fixMdHtml_to_Html("_site/"+aPage+"/"+fn_all[f]+".html") )				// Xyzzy - make this into a func.
					if err != nil { panic(err) }
					defer func() {														// close fo on exit and check for its returned error
						if err := fo.Close(); err != nil {
							panic(err)
						}
					}()

					pass1Data.Td = templateData
					// dumpVar ( pass1Data )
					err = t.ExecuteTemplate(fo, aPage, pass1Data)
					if err != nil { panic(err) }

					// Save the processed template for other files like index.html =============================================================================
					// b = prefixPostfix ( "{{define \"body\"}}\n", b, "\n{{end}}\n" ) -- Got the body - so just need to use it in "Index"
					if fn_all[f] == mostRecentFnAll {
fmt.Printf ( "aPage[%s] fn_all[%d] = ->%s<- at last_pos=%d\n", aPage, f, fn_all[f], last_pos );
						doc := bytes.NewBuffer(make([]byte, 0))
						err = t.ExecuteTemplate(doc, aPage+".fi", pass1Data)						// xyzzy
						// err = t.ExecuteTemplate(doc, aPage, pass1Data)
						if err != nil { panic(err) }
						post_b := doc.String()
						allOfThePosts[last_pos].BodyMostRec = template.HTML(post_b)
						// ---- pos needs to be identified.
						// fmt.Printf ( "->%s<-\n", post_b );
						// fmt.Printf ( "->%d<-\n", last_pos );
					}

				}
			}

			g_IncludePath = "{{.Dir}}|./tmpl"
			templateData.IncludePath = "{{.Dir}}|./tmpl"
		}
	}

	if debug {
		fmt.Printf ( "final Data =\n" );
		dumpVar ( pass1Data )
	}

	// Part 5 - Do the global "Makefile" - "make push_up"
	arg0[0] = "make"
	arg0[1] = "finialize"
	fmt.Printf ( "make finialize\n" );
	executeBashCmdSlice ( arg0 )

	arg0[1] = "push_up"
	fmt.Printf ( "make push_up\n" );
	executeBashCmdSlice ( arg0 )

}



