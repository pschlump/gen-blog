/*

*/
package main

import (
	"./mux"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"strconv"
)
	// "github.com/nu7hatch/gouuid"
    // _ "github.com/lib/pq"
	//"code.google.com/p/gorilla/mux"			-- Pulled in locally and modified
	//"log"
	//"reflect"

const serverPort = ":8764"

const base64GifPixel = "R0lGODlhAQABAIAAAP///wAAACwAAAAAAQABAAACAkQBADs="

var fo *os.File
var fx *os.File

var q_cnt int

// -------------------------------------------------- Support Funcs  --------------------------------------------------

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

func  uriToStringMap ( req *http.Request ) ( m url.Values ) {
	u, _ := url.ParseRequestURI(req.RequestURI)
	m, _ = url.ParseQuery(u.RawQuery)
	return
}

func validateQueryParams ( m url.Values, v interface{} ) ( n url.Values, err error ) {
	n = m
	err = nil
	return
}











// -------------------------------------------------- Handlers --------------------------------------------------

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
	q := req.RequestURI

	io.WriteString(res,jsonP("{\"status\":\"success\",\"query\":\""+q+"\"}",res,req))
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

//	mux.HandleFunc("/api/table/{name}", respHandlerTableGet ).Methods("GET")
//	mux.HandleFunc("/api/table/{name}", respHandlerTablePut ).Methods("PUT")
//	mux.HandleFunc("/api/table/{name}", respHandlerTablePost).Methods("POST")
//	mux.HandleFunc("/api/table/{name}", respHandlerTableDel ).Methods("DEL")
//	mux.HandleFunc("/api/table/{name}", respHandlerTableHead).Methods("HEAD")
func respHandlerTableGet(res http.ResponseWriter, req *http.Request) {
    res.Header().Set("Content-Type","application/json")
    io.WriteString(res,"{\"status\":\"success\",\"method\":\"post\"}")
}
func respHandlerTablePut(res http.ResponseWriter, req *http.Request) {
    res.Header().Set("Content-Type","application/json")
    io.WriteString(res,"{\"status\":\"success\",\"method\":\"post\"}")
}

func respHandlerTablePost(res http.ResponseWriter, req *http.Request) {
    res.Header().Set("Content-Type","application/json")
    io.WriteString(res,"{\"status\":\"success\",\"method\":\"post\"}")
}

func respHandlerTableDel(res http.ResponseWriter, req *http.Request) {
    res.Header().Set("Content-Type","application/json")
    io.WriteString(res,"{\"status\":\"success\",\"method\":\"post\"}")
}

func respHandlerTableHead(res http.ResponseWriter, req *http.Request) {
    res.Header().Set("Content-Type","application/json")
    io.WriteString(res,"{\"status\":\"success\",\"method\":\"post\"}")
}



func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Silly World<br>"))
}
func homeRobotsTxt(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("User-agent: *\nDisallow: /api/\nDisallow: /static/img/\n"))
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
	mux.HandleFunc("/robots.txt", homeRobotsTxt)
	mux.HandleFunc("/static/img/Blank.gif", respHandler).Methods("GET","POST")
	mux.HandleFunc("/static/img/blank.gif", respHandlerNoJs).Methods("GET","POST")
	mux.HandleFunc("/api/swapLogFile/{seq}", respHandlerSwapLogFile).Methods("GET","POST")
	mux.HandleFunc("/api/registerEmail", respHandlerRegEmail).Methods("GET","POST")
	mux.HandleFunc("/api/deRegisterEmail", respHandlerDeRegEmail).Methods("GET","POST")
	mux.HandleFunc("/api/grabFeedback", respHandlerGrabFeedback).Methods("GET")
	mux.HandleFunc("/api/status", respHandlerStatusGet).Methods("GET")
	mux.HandleFunc("/api/status", respHandlerStatusPost).Methods("POST")
	mux.HandleFunc("/api/status", respHandlerStatusHead).Methods("HEAD")


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

