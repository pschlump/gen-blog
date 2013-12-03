
/*

	Server
		0. Open log file 
		1. Use "arg1" as output dir for log fles
		2. If file curently exists move it to file.seq
	Client
		0. add in JS/HTML code to client
		1. Change JS code to have URL -- Via template would be best

	Server - Write node.js code to read log files and load in database.
	Implement a robots.txt that prevents "get" on Blank.gif etc.

-- later -- Cron job to do work.
	1. get coonected to database -- Log to file instead
	2. create tables with udpate triggers
	4. do "insert" into database with stuff

*/

package main

import (
	"./mux"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"./gouuid"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"strconv"
)
	// "github.com/nu7hatch/gouuid"
	//"code.google.com/p/gorilla/mux"
	//"log"

const serverPort = ":8764"

const base64GifPixel = "R0lGODlhAQABAIAAAP///wAAACwAAAAAAQABAAACAkQBADs="

var fo *os.File
var fx *os.File

func dumpVar ( v interface{} ) {
	// s, err := json.Marshal ( v )
	s, err := json.MarshalIndent ( v, "", "\t" )
	if ( err != nil ) {
		fmt.Printf ( "Error: %s\n", err )
	} else {
		fmt.Printf ( "%s\n", s )
	}
}

func jsonP ( s string, res http.ResponseWriter, req *http.Request ) string {
	u, _ := url.ParseRequestURI(req.RequestURI)
	m, _ := url.ParseQuery(u.RawQuery)
	callback := m.Get("callback")
	if ( callback != "" ) {
		res.Header().Set("Content-Type","application/javascript")				// For JSONP
		return fmt.Sprintf("%s(%s);",callback,s)
	} 
	return s
}





func respHandler(res http.ResponseWriter, req *http.Request) {
    res.Header().Set("Content-Type","image/gif")
    output,_ := base64.StdEncoding.DecodeString(base64GifPixel)
    io.WriteString(res,string(output))
}

func respHandlerNoJs(res http.ResponseWriter, req *http.Request) {
    res.Header().Set("Content-Type","image/gif")
    output,_ := base64.StdEncoding.DecodeString(base64GifPixel)
    io.WriteString(res,string(output))
}

func respHandlerGrabFeedback(res http.ResponseWriter, req *http.Request) {
    res.Header().Set("Content-Type","image/gif")
    output,_ := base64.StdEncoding.DecodeString(base64GifPixel)
    io.WriteString(res,string(output))
}

func respHandlerRegEmail(res http.ResponseWriter, req *http.Request) {
    //res.Header().Set("Content-Type","application/javascript")				// For JSONP
    //io.WriteString(res,"{\"status\":\"success\"}")
    res.Header().Set("Content-Type","image/gif")
    output,_ := base64.StdEncoding.DecodeString(base64GifPixel)
    io.WriteString(res,string(output))
	//fmt.Printf ( "\tGot Status Request\n" );
	//fmt.Printf("\tLOG: %s %s %s %v\n", req.RemoteAddr, req.Method, req.URL, req.Header)
}

func respHandlerDeRegEmail(res http.ResponseWriter, req *http.Request) {
    res.Header().Set("Content-Type","application/javascript")				// For JSONP
    io.WriteString(res,"{\"status\":\"success\"}")
	//fmt.Printf ( "\tGot Status Request\n" );
	//fmt.Printf("\tLOG: %s %s %s %v\n", req.RemoteAddr, req.Method, req.URL, req.Header)
}

func respHandlerStatusGet(res http.ResponseWriter, req *http.Request) {
    res.Header().Set("Content-Type","application/json")

	// fmt.Printf ( "URL = %v\n", req.URL );
	// dumpVar ( req );

	io.WriteString(res,jsonP("{\"status\":\"success\"}",res,req))
	fmt.Printf ( "***Got Status Request\n" );
	//fmt.Printf("\tLOG: %s %s %s %v\n", req.RemoteAddr, req.Method, req.URL, req.Header)
}
func respHandlerStatusPost(res http.ResponseWriter, req *http.Request) {
    res.Header().Set("Content-Type","application/json")
    io.WriteString(res,"{\"status\":\"success\",\"method\":\"post\"}")
	fmt.Printf ( "***Got Status Request - Post\n" );
	//fmt.Printf("\tLOG: %s %s %s %v\n", req.RemoteAddr, req.Method, req.URL, req.Header)
}

func respHandlerStatusHead(res http.ResponseWriter, req *http.Request) {
    res.Header().Set("Content-Type","application/json")
    io.WriteString(res,"{\"status\":\"success\",\"method\":\"head\"}")
	fmt.Printf ( "***Got Status Request - Head\n" );
	//fmt.Printf("\tLOG: %s %s %s %v\n", req.RemoteAddr, req.Method, req.URL, req.Header)
}


func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Silly World<br>"))
}

const debug = false

func dumpURL ( s string, req *http.Request ) {
	if debug {
		fmt.Printf ( "%s\n", s );
		fmt.Printf ( "\treq.URL.Scheme=%s\n", req.URL.Scheme)
		fmt.Printf ( "\treq.URL.Host=%s\n", req.URL.Host)
		fmt.Printf ( "\treq.URL.Path=%s\n", req.URL.Path)
		fmt.Printf ( "\treq.URL.RawQuery=%s\n", req.URL.RawQuery)
		fmt.Printf ( "\treq.URL.Fragment=%s\n", req.URL.Fragment)
	}
}

func respHandlerSwapLogFile(res http.ResponseWriter, req *http.Request) {
	var err error

	if err := fo.Close(); err != nil {
		fmt.Printf("Error: %v\n",err)
	}
	if err := fx.Close(); err != nil {
		fmt.Printf("Error: %v\n",err)
	}

	dumpURL ( "respHandlerSwapLogFile", req );

	// get seq # from cmd line
	seq := mux.Vars(req)["seq"]
	// fmt.Printf ( "seq=%s\n", seq)
	// rename files
	os.Rename ( "log/alog.log", "log/alog.log."+seq )
	os.Rename ( "log/xlog.log", "log/xlog.log."+seq )

	fo, err = os.Create("log/alog.log")
	if err != nil { panic(err) }
    // close fo on exit and check for its returned error
    defer func() {
        if err := fo.Close(); err != nil {
            panic(err)
        }
    }()
	fx, err = os.Create("log/xlog.log")
	if err != nil { panic(err) }
    // close fo on exit and check for its returned error
    defer func() {
        if err := fx.Close(); err != nil {
            panic(err)
        }
    }()
    res.Header().Set("Content-Type","application/javascript")				// For JSONP
    io.WriteString(res,"{\"status\":\"success\"}")
}










// ------------------------------------- log code -----------------------------------------------------------------
//  From https://gist.github.com/cespare/3985516 

const ApacheFormatPattern = "%s %v %s %v %d %v\n"

type ApacheLogRecord struct {
	http.ResponseWriter

	ip                    string
	time                  time.Time
	method, uri, protocol string
	status                int
	responseBytes         int64
	elapsedTime           time.Duration
}

func (r *ApacheLogRecord) Log(out io.Writer) {
	timeFormatted := r.time.Format("02/Jan/2006 03:04:05")
	requestLine := fmt.Sprintf("%s %s %s", r.method, r.uri, r.protocol)
	fmt.Fprintf(out, ApacheFormatPattern, r.ip, timeFormatted, requestLine, r.status, r.responseBytes, r.elapsedTime.Seconds())
//	fmt.Fprintf(out,"\tLOG: %s %v %s %v %d %v\n", r.ip, timeFormatted, requestLine, r.status, r.responseBytes, r.elapsedTime.Seconds())
//	fmt.Fprintf(out,"\tLOG: %s %s %s %v\n", req.RemoteAddr, req.Method, req.URL, req.Header)
}

/*
func getInt( m , name string ) rv int {
	rv, err := strconv.Atoi( m.Get(name) )
	if err != nil {
		rv = 0
	}
	return
}

func getString( m url.Values, name string ) rv string {
	rv, err := m.Get(name)
	if err != nil {
		rv = ""
	}
	return
}
*/

func (r *ApacheLogRecord) XLog(out io.Writer,req *http.Request) {

	var s_cookie string
	cookie, e0 := req.Cookie("blog-2c-why-cookie")
	if e0 == nil {
		s_cookie = cookie.String()
	} else {
		s_cookie = ""
	}

	u, _ := url.ParseRequestURI(r.uri)
	// clear_gif.src = 'http://localhost:8764/api/Blank.gif?key=blog2cwhy&c="+c+"&sw="+screen.width+"&sh="+screen.height+"&cd="+screen.colorDepth;
	m, _ := url.ParseQuery(u.RawQuery)
	/*
	sh := getInt(m,"sh")
	sw := getInt(m,"sw")
	cd := getInt(m,"cd")
	key := getString(m,"key")
	*/

	sh, e1 := strconv.Atoi( m.Get("sh") )
	if e1 != nil {
		sh = 0
	}
	sw, e2 := strconv.Atoi( m.Get("sw") )
	if e2 != nil {
		sw = 0
	}
	cd, e3 := strconv.Atoi( m.Get("cd") )
	if e3 != nil {
		cd = 0
	}
	if s_cookie == "" {
		s_cookie = m.Get("c")
		/*
		s_cookie = getString(m,"c")
		*/
	}
	key := m.Get("key")
	ref := m.Get("ref")
	tz := m.Get("tz")
	url := m.Get("u")

// Parse out "file" component "Blank.gif" or "blank.gif"

	x := map[string]interface{}{
		"ip": r.ip,
		"timeFormatted": r.time.Format(time.RFC3339Nano),
		"method": r.method,
		"uri": r.uri,
		"protocol": r.protocol,
		"status": r.status,
		"responseBytes": r.responseBytes,
		"elapsedTimeString": fmt.Sprintf ( "%f", r.elapsedTime.Seconds()),
		"referer": ref,
		"cookie": s_cookie,
		"user_agent": req.UserAgent(),
		"screen_height": sh,
		"screen_width": sw,
		"colorDepth": cd,
		"userKey": key,
		"timeZoneOffset": tz,
		"url": url,
	}

	s, e4 := json.MarshalIndent ( x, "", "\t" )
	if e4 != nil {
		fmt.Fprintf(out,"Error: %s\n", e4 );
	} else {
		fmt.Fprintf(out,"Data: %s\n", s );
	}
}

func (r *ApacheLogRecord) Write(p []byte) (int, error) {
	written, err := r.ResponseWriter.Write(p)
	r.responseBytes += int64(written)
	return written, err
}

func (r *ApacheLogRecord) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

type ApacheLoggingHandler struct {
	handler http.Handler
	out     io.Writer
	xout    io.Writer
}

func NewApacheLoggingHandler(handler http.Handler, out io.Writer, xout io.Writer) http.Handler {
	return &ApacheLoggingHandler{
		handler: handler,
		out:     out,
		xout:    xout,
	}
}

func (h *ApacheLoggingHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	if colon := strings.LastIndex(clientIP, ":"); colon != -1 {
		clientIP = clientIP[:colon]
	}

	record := &ApacheLogRecord{
		ResponseWriter: rw,
		ip:             clientIP,
		time:           time.Time{},
		method:         r.Method,
		uri:            r.RequestURI,
		protocol:       r.Proto,
		status:         http.StatusOK,
		elapsedTime:    time.Duration(0),
	}

	dumpURL ( "Top ServeHTTP", r );
	startTime := time.Now()
	h.handler.ServeHTTP(record, r)
	finishTime := time.Now()
	dumpURL ( "Bot ServeHTTP", r );

	record.time = finishTime.UTC()
	record.elapsedTime = finishTime.Sub(startTime)

	record.Log(h.out)
	record.XLog(h.xout,r)
}

// ------------------------------------- main -----------------------------------------------------------------
// os.Args

func main() {

	var err error

// open output file: a-log.log
    // fo, err := os.Create("log/alog.log")
    fo, err = os.OpenFile("log/alog.log", os.O_RDWR|os.O_APPEND, 0660)
    if err != nil {
		fo, err = os.Create("log/alog.log")
		if err != nil { panic(err) }
	}
    // close fo on exit and check for its returned error
    defer func() {
        if err := fo.Close(); err != nil {
            panic(err)
        }
    }()

// open output file: x-log.log
    fx, err = os.OpenFile("log/xlog.log", os.O_RDWR|os.O_APPEND, 0660)
    if err != nil {
		fx, err = os.Create("log/xlog.log")
		if err != nil { panic(err) }
	}
    // close fo on exit and check for its returned error
    defer func() {
        if err := fx.Close(); err != nil {
            panic(err)
        }
    }()


	mux := mux.NewRouter()

	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/index.html", homeHandler)
	mux.HandleFunc("/static/img/Blank.gif", respHandler).Methods("GET","POST")
	mux.HandleFunc("/static/img/blank.gif", respHandlerNoJs).Methods("GET","POST")
	mux.HandleFunc("/api/swapLogFile/{seq}", respHandlerSwapLogFile).Methods("GET","POST")
	mux.HandleFunc("/api/registerEmail", respHandlerRegEmail).Methods("GET","POST")
	mux.HandleFunc("/api/deRegisterEmail", respHandlerDeRegEmail).Methods("GET","POST")
	mux.HandleFunc("/api/grabFeedback", respHandlerGrabFeedback).Methods("GET")
	mux.HandleFunc("/api/status", respHandlerStatusGet).Methods("GET")
	mux.HandleFunc("/api/status", respHandlerStatusPost).Methods("POST")
	mux.HandleFunc("/api/status", respHandlerStatusHead).Methods("HEAD")
	/*
			// 1. "observed" URL
			// 2. "visited" by this user
			// 3. "pullData" for this user from server
			// 4. paint the existing data
			// 3. onClick - #edit-y-bar - create a sticky
			//    1. Locate postion to attach - relative pos from element
			//    2. Allow drag/drop of item at height of 501
			//    3. Grab focus for title - show time
			//    4. saveContent - when (submit button) ( every 200ms if delta text )
			// 4. onClick of #edit-top
			//    0. saveContent 
			//    1. change locatio back to /edit-me.html
			// -----
			// x. Add a "E" button to edit in the visual editor for markdown -- Edit the .md
			// x. Add a "D" button to mark as done with notes/comments edits
	// See line 221 for Url Parameters, ?callback=XXXXX - if has that then wrap response in JSON-P callback.
	mux.HandleFunc("/api/observedPage", respHandlerObservedPageGet).Methods("GET")  		// A page was seen -- add URL to list of urls for master
	mux.HandleFunc("/api/visitedPage", respHandlerVisitedPageGet).Methods("GET")			// An editor has clicked on a page to visit it.  Mark it as visited by that editor.
	mux.HandleFunc("/api/assignPageTo", respHandlerAssignPageToGet).Methods("GET")			// Master has assigned the page to an editor to work on it
	mux.HandleFunc("/api/pullDataFor", respHandlerPullDataForGet).Methods("GET")			// Pull all the data for this user ( master is union of all ) - Input URL + User - Output JSON of all notes for page
	mux.HandleFunc("/api/pullDataTopUser", respHandlerPullDataTopUserGet).Methods("GET")	// Pull all the data for painting the top edit page (vititeds, done etc for all URLs, # of notes per url etc)
	mux.HandleFunc("/api/saveDataPage", respHandlerSaveDataPageGet).Methods("GET")			// Save the data for a page (all the notes and positions for a single user)
	mux.HandleFunc("/api/saveOneNote", respHandlerSaveOneNoteGet).Methods("GET")			// Save the data for a single edited/new note (input has pageID)
	mux.HandleFunc("/api/markEditorDone", respHandlerMarkEditorDoneGet).Methods("GET")		// Mark that an editor is done looking at / reviewing a page - they cliked the "done" button for page "D" - has Y/N done flag for "undone"
	mux.HandleFunc("/api/fetchMDPage", respHandlerFetchMDPageGet).Methods("GET")			// Get the raw ".md" data for a page
	mux.HandleFunc("/api/saveMDPage", respHandlerSaveMDPageGet).Methods("GET")				// Save modified ".md" data for a page - do a git operation to save
	*/
	mux.PathPrefix("/static/").Handler(http.FileServer(http.Dir(".")))

	loggingHandler := NewApacheLoggingHandler(mux, fo, fx)
	err = http.ListenAndServe(serverPort, loggingHandler)
	//server := &http.Server{
	//	Addr:    serverPort,
	//	Handler: loggingHandler,
	//}
	//err = server.ListenAndServe()
	if  err != nil {
		fmt.Printf ("Error from server %v\n", err )
	}
}


func old_respHandler(res http.ResponseWriter, req *http.Request) {
    res.Header().Set("Content-Type","image/gif")
    output,_ := base64.StdEncoding.DecodeString(base64GifPixel)
	etag, _ := uuid.NewV4()
	res.Header().Set("Etag",etag.String())
	res.Header().Set("If-None-Match",etag.String())
    io.WriteString(res,string(output))
	//fmt.Printf ( "\tGot Request for 'Blank.gif' - indicates have JS\n" )
	//fmt.Printf("\tLOG: %s %s %s %v\n", req.RemoteAddr, req.Method, req.URL, req.Header)
}
