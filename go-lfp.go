// xLog file processor
/*
	Log File Processor go-lfp.go
		.sh
		0. read in the "seq" for the log files - find last one add one
		1. Ask :8764 to swap log files (get request) - get request
		2. grep for '"url"'
		3. go/proc
			1. read in freq.json
			2. read in new data and accumulate
			3. write freq.json (plus make backup of old with "seq" - from old)
		
*/

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
    "io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const rootUrl = "http://localhost:8764"

type  UsageData struct {
	Base		string
	Seq			string
	Usage		[]UsageStats
}

type  UsageStats struct {
	Url			string
	Count		int
}

var globalUsageData UsageData
var debug bool = true

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

func varToJson ( v interface{} ) string {
	s, err := json.MarshalIndent ( v, "", "\t" )
	if ( err != nil ) {
		return fmt.Sprintf ( "{ \"error\": \"%s\" }\n", err )			// Xyzzy - may need to string encode/repalce " with \\" in err
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

func incrementCount ( url string ) {
	if debug { fmt.Printf ( "incrementCount(url=%s)\n", url ) }
	var re_s1 = regexp.MustCompile(".*\"url\"[ ]*:.*http://[^/]*//*")				// get the seq #
	var re_s2 = regexp.MustCompile("\",.*")
	url = re_s1.ReplaceAllString(url, "")
	if debug { fmt.Printf ( "incrementCount(url=%s) after cleanup head\n", url ) }
	url = re_s2.ReplaceAllString(url, "")
	if debug { fmt.Printf ( "incrementCount(url=%s) after cleanup tail\n", url ) }

	for k, v := range globalUsageData.Usage {
		if v.Url == url {
			globalUsageData.Usage[k].Count++
			return
		}
	}
	nnn := UsageStats{
		Url: url,
		Count: 1,
	}
	globalUsageData.Usage = append ( globalUsageData.Usage, nnn )

}

func main() {

	var next_seq_s string

	// --- Read and parse the current freq.json file
	usage, e8 := ioutil.ReadFile("./freq.json")
	if e8 != nil {
		fmt.Printf("File error: %v\n", e8)
		os.Exit(1)
	}
	json.Unmarshal(usage, &globalUsageData)


	// --- Get the next sequence by listing files in log dir and using that for seq.
	filenames, _  := getFilenames ( "./log" )
	if debug { fmt.Printf ( "Original filenames = %v\n", filenames ) }
	filenames = filterArray( `xlog\.log\.`, filenames )	// flter file names for "xlog.log" files!
	if debug { fmt.Printf ( "Filtered filenames = %v\n", filenames ) }
	sort.Strings(filenames)
	if debug { fmt.Printf ( "Sorted filenames = %v\n", filenames ) }

	last_fn := filenames[len(filenames)-1]					// Pick last file
	if debug { fmt.Printf ( "last filenames = %s\n", last_fn ) }
	var re = regexp.MustCompile("^xlog.log.")				// get the seq #
	seq_s := re.ReplaceAllString(last_fn, "")
	if debug { fmt.Printf ( "seq_s = %s\n", seq_s ) }
	seq_n, err := strconv.Atoi( seq_s )
	if debug { fmt.Printf ( "seq_n = %d\n", seq_n ) }
	next_seq_s = fmt.Sprintf ( "%06d", seq_n+1 )					// Add 1 and format
	if debug { fmt.Printf ( "next_seq_s = %s\n", next_seq_s ) }
	globalUsageData.Seq = next_seq_s								// For later - backup of seq data - use this if no *.000000 files in ./log

	// Save a copy of the freq data in backup with seq (old) since it matches with (old) data.
	ioutil.WriteFile("./bak/freq.json."+seq_s, usage, 0644 )

	if debug { fmt.Printf ( "get ->%s<-\n", rootUrl+"/api/swapLogFile?seq="+next_seq_s) }
	// res, err := http.Get(rootUrl+"/api/swapLogFile?seq="+next_seq_s)		// Ask for log file switch
	res, err := http.Get(rootUrl+"/api/swapLogFile/"+next_seq_s+"?seq2="+next_seq_s+"#frag12")		// Ask for log file switch
	if err != nil {
		log.Fatal(err)
	}
	if debug { fmt.Printf ( "res = %v\n", res ) }


	// Open log file (new seq)
    fi, err := os.Open("./log/xlog.log."+next_seq_s)
    // fi, err := os.Open("./log/xlog.log."+seq_s)
    if err != nil { panic(err) }
    defer func() {
        if err := fi.Close(); err != nil {
            panic(err)
        }
    }()

	// Read and pares it 1 line at a time - convert JSON iobuffer.ReadString(...)
	var url = regexp.MustCompile("\"url\"[ ]*:.*http")				// get the seq #
	scanner := bufio.NewScanner(fi)
	for scanner.Scan() {
		// fmt.Println(scanner.Text()) // Println will add back the final '\n'
		s := scanner.Text()
		// Extract "url" component - if match with "url"j
		if  url.MatchString( s ) {
			// Update/Insert/Add in globalUsageData -- addToFreq ( url )
			incrementCount ( s )
		}
	}

	// write new freq.json file
	out := varToJson ( globalUsageData )
	if debug { fmt.Printf ( "->%s<-\n", out ); }
	ioutil.WriteFile("freq.json", []byte(out), 0644 )

}

