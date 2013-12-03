
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

ToDo:
	1. Add connect to Redis
	2. Add /api/table/{name} calls
	3. Add where clause as parse tree
	4. Add caching of results in Redis


Future / Things you can help with
	Notifications: LISTEN/NOTIFY
http://bjorngylling.com/2011-04-13/postgres-listen-notify-with-node-js.html

http://redneckbeard.github.io/gadget/ - famework in Go

https://github.com/codegangsta/cli -- CLI to do go/git like subcommands

-- later -- Cron job to do work.
	1. get coonected to database -- Log to file instead
	2. create tables with udpate triggers
	4. do "insert" into database with stuff


http://blog.golang.org/error-handling-and-go -- Error handlingin
./net/http/server.go:1723 - Redirect
./net/http/server.go:1230 - Redirect

path.Clean()

in ./net/server.go -----  func cleanPath(p string) string {

*/

package main

import (
    _ "./pq"
    "database/sql/driver"
    "database/sql"
	"./mux"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"./gouuid"
	"io"
    "io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"strconv"
	"runtime"
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

func connectToDb() (*sql.DB) {
	q_cnt = 0
    db, err := sql.Open("postgres", "user=postgres password=f1ref0x2 dbname=test sslmode=disable")
	if err != nil { panic(err) }
	db.SetMaxIdleConns(5)
	// defer db.Close()
	return db
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




func validateUser ( m url.Values ) ( n url.Values, err error ) {
	var id string
	var master string
	var color string
	var bar string
	var text string

	_, ok := m["user"];
	if ! ok {
		err = errors.New( "Missing Parameter:user" );
		fmt.Printf ( "Request requreis a 'user' parameter.  It was not suplied.\n" );
		return
	}
	_, ok = m["apikey"];
	if ! ok {
		err = errors.New( "Missing Parameter:apikey" );
		fmt.Printf ( "Request requreis a 'apikey' parameter.  It was not suplied.\n" );
		return
	}

	n = m
    Rows, err := db.Query("select \"id\", \"master\", \"color\", \"bar\", \"text\" from \"e_user\" where \"username\" = $1 and \"apikey\" = $2", m["user"][0], m["apikey"][0] )
	q_cnt++
	// defer Rows.Close()


	if err == driver.ErrBadConn {
fmt.Printf ( "****************************************** error occured ****************************************** %d\n", q_cnt );
	}

	// fmt.Printf ( "err = %v\n", err )
	switch {
    case err == sql.ErrNoRows:
		// fmt.Printf ( "ValidateUser no rows\n" );
		return
        //   log.Fatal(err)				// xyzzy log it
    case err != nil:
		// fmt.Printf ( "ValidateUser d.b. error, %v\n", err );
		return
        //   log.Fatal(err)				// xyzzy log it
    default:
		// fmt.Printf ( "ValidateUser got X row\n" );
		n_row := 0
		for Rows.Next() {
			n_row++
			err = Rows.Scan(&id,&master,&color,&bar,&text)
			n.Add ( "user_id", id )
			n.Add ( "user_master", master )
			n.Add ( "user_color", color )
			n.Add ( "user_bar", bar )
			n.Add ( "user_text", text )
		}
		if n_row != 1 {
			err = sql.ErrNoRows				// xyzzy log it
		}
    }
	// fmt.Printf ( "at end ov ValidateUser - n=" )
	// dumpVar ( n )
	return
}

// -------------------------------------------------------------------------------------------------
// Run a database query, return the rows.  Handle errors.
// -------------------------------------------------------------------------------------------------
func sel ( res http.ResponseWriter, req *http.Request, db *sql.DB, q string, data ...interface{} ) ( Rows *sql.Rows, err error ) {
	fmt.Printf ( "Query (%s) with data:", q )
	dumpVar ( data )
    Rows, err = db.Query(q, data... )
	q_cnt++
	// defer Rows.Close()
	if err == driver.ErrBadConn {
		fmt.Printf ( "****************************************** error occured ****************************************** %d\n", q_cnt );
	}
    if err != nil {

		_, file, line, _ := runtime.Caller(1)
		fmt.Printf ( "Database error (%v) at %s:%d\n", err, file, line ) // xyzzy - need to escape quotes and pass this back in JSON - what about '\' and '''' - encode those?
																			// xyzzy - really should log this
		detail := fmt.Sprintf ( "%v", err )
		detail = strings.Replace(detail,"\"","\\\"",-1)
		io.WriteString(res,jsonP(fmt.Sprintf("{\"status\":\"error\",\"code\":\"2\",\"msg\":\"Database error\",\"file\":\"%s\",\"line\":%d,\"detail\":\"%s\"}",file,line,detail),res,req))

    }
	return
}









const ISO8601 = "2006-01-02T15:04:05.99999Z07:00"

// -------------------------------------------------------------------------------------------------
// Rows to JSON -- Go from a set of "rows" returned by db.Query to a JSON string.
// -------------------------------------------------------------------------------------------------
func rowsToJson ( rows *sql.Rows ) string {

	var finalResult   []map[string]interface{}
	var oneRow         map[string]interface{}

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error())
	}
	length := len(columns)

	// Make a slice for the values
	values := make([]interface{}, length)

	// rows.Scan wants '[]interface{}' as an argument, so we must copy the
	// references into such a slice
	// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
	scanArgs := make([]interface{}, length)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Fetch rows
	for rows.Next() {
		oneRow = make(map[string]interface{}, length)
		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}

		// Print data
		for i, value := range values {
			switch value.(type) {
			case nil:
				// fmt.Println("n", columns[i], ": NULL")
				oneRow[columns[i]] = nil

			case []byte:
				// fmt.Println("s", columns[i], ": ", string(value.([]byte)))
				oneRow[columns[i]] = string(value.([]byte))

			case int64:
				// fmt.Println("i", columns[i], ": ", value)
				oneRow[columns[i]] = fmt.Sprintf ( "%v", value )

			case float64:
				//fmt.Println("f", columns[i], ": ", value)
				oneRow[columns[i]] = fmt.Sprintf ( "%v", value )

			case bool:
				//fmt.Println("b", columns[i], ": ", value)
				oneRow[columns[i]] = fmt.Sprintf ( "%v", value )

			case string:
				// fmt.Println("S", columns[i], ": ", value)
				oneRow[columns[i]] = fmt.Sprintf ( "%s", value )

			// xyzzy - there is a timeNull structure in the driver - why is that not returned?  Maybee it is????
			case time.Time:
				// 2013-10-09 18:45:07.189352 +0000 +0000 
				//	s := t.Format ( "2006-01-02T15:04:05Z07:00"  )
				ss := fmt.Sprintf ( "%v", value );
				// fmt.Printf ( "ss=%s\n", ss );
				t, err := time.Parse( "2006-01-02 15:04:05.999999 -0700",  ss[0:32])
				if err != nil {
					// fmt.Printf ( "err=%v\n", err );
					oneRow[columns[i]] = nil
				} else {
					//fmt.Println("t", columns[i], ": ", t.Format(ISO8601) )
					oneRow[columns[i]] = t.Format(ISO8601) 
				}

			default:
				// fmt.Println("r", columns[i], ": ", value)
				oneRow[columns[i]] = fmt.Sprintf ( "%v", value )
			}
			//fmt.Printf("\nType: %s\n", reflect.TypeOf(value))
		}
		// fmt.Println("-----------------------------------")
		finalResult = append ( finalResult, oneRow )
	}
	s, _ := json.MarshalIndent ( finalResult, "", "\t" )
	return string(s)
}









// -------------------------------------------------- New --------------------------------------------------
type Validation struct {
	Required		bool
	Optional		bool
	Type			string
	Min_len			int
	Max_len			int
	Min				int
	Max				int
	Default			string
}

type SQLOne struct {
	F		string
	P		[]string
	Query	string
	Valid	map[string]Validation
	Nokey			bool
}
var SQLCfg map[string]SQLOne

//------------------------------------------------------------------------------------------------
// read in the configuration file for the queries / function calls that can be handled by
// respHandlerSQL
//
// Notes:
//	1. Should add stuff for generic queries
//	2. Figure out how to do an /api/table/<name>?params
//	3. Figure out how to do full CRUD
//------------------------------------------------------------------------------------------------
func readInSQLConfig(path string) map[string]SQLOne {
	var jsonData map[string]SQLOne
	file, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf ( "error (135): %v\n", err )
		return jsonData
	}
	err = json.Unmarshal( file, &jsonData)
	if err != nil {
		fmt.Printf ( "error (140): %v\n", err )
	}
	return jsonData
}

//------------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------------
func respHandlerSQL(res http.ResponseWriter, req *http.Request, cfgTag string) {
	h := SQLCfg[cfgTag]										// get configuration						// xyzzy - what if config not defined for this item at this point!!!!

	// fmt.Printf ( "at top of respHandleSQL - h=" )
	// dumpVar ( h )

    res.Header().Set("Content-Type","application/json")		// set default reutnr type

	m := uriToStringMap ( req )								// pull out the ?name=value params

	m, err := validateQueryParams ( m, h.Valid )			// Validate them
	if err != nil {
		io.WriteString(res,jsonP("{\"status\":\"error\",\"code\":\"3\",\"msg\":\"Invalid Query Parameters\"}",res,req))
		return
	}

	if ! h.Nokey {
		m, err = validateUser ( m )								// see if this is a valid user
		if err != nil {
			io.WriteString(res,jsonP("{\"status\":\"error\",\"code\":\"1\",\"msg\":\"Invalid API Key\"}",res,req))
			return
		}
	}

	data := make([]interface{}, len(h.P))					// organize data for call to d.b.
	for i := 0; i < len(h.P); i++ {
		kk := h.P[i];
		_, ok := m[kk];
		if ok {
			data[i] = m[h.P[i]][0]
		} else {
			fmt.Printf ( "******* Error ******* - Missing data for %s, Empty string used!\n", kk );
			data[i] = "";
		}
	}

	if h.F != "" {											// make call to d.b.
		// fmt.Printf ( "Dong F (%s)\n", h.F )
		Rows, err := sel ( res, req, db, h.F, data... )
		if err != nil { return }
		if h.Query == "" {
			io.WriteString(res,jsonP("{\"status\":\"success\"}",res,req))
		}
		Rows.Close()
	} 
	if h.Query != "" {
		// fmt.Printf ( "Dong Query (%s)\n", h.Query )
		Rows, err := sel ( res, req, db, h.Query, data... )
		if err != nil { return }
		s := rowsToJson(Rows)
		fmt.Printf ( "Returned Data: %s\n", s )
		io.WriteString(res,jsonP(s,res,req))
		// Rows.Close()
	}
}





// ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// Create the "closure" fucntion that will save passed data for later and return a 
// function bound to the passed data.
func getSqlCfgHandler(name string) ( func (res http.ResponseWriter, req *http.Request) ) {
	return func (res http.ResponseWriter, req *http.Request) {
		respHandlerSQL(res,req,name)
	}
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
// os.Args
var db *sql.DB

func main() {

	var err error

	SQLCfg = readInSQLConfig("sql-cfg.json")

	// dumpVar ( SQLCfg )
	// os.Exit(0)

	db = connectToDb() 

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


// xyzzy - xyzzy - could pull all of these from the sql-*.json file - and then track changes to file.

	mux.HandleFunc("/api/observedPage", getSqlCfgHandler("respHandlerObservedPageGet")).Methods("GET")			// A page was seen -- add URL to list of urls for master
	mux.HandleFunc("/api/observedPage2", getSqlCfgHandler("respHandlerObservedPage2Get")).Methods("GET")		// A page was seen -- add URL to list of urls for master
	mux.HandleFunc("/api/visitedPage", getSqlCfgHandler("respHandlerVisitedPageGet")).Methods("GET")			// An editor has clicked on a page to visit it.  Mark it as visited by that editor.
	mux.HandleFunc("/api/assignPageTo", getSqlCfgHandler("respHandlerAssignPageToGet")).Methods("GET")			// Master has assigned the page to an editor to work on it
	mux.HandleFunc("/api/pullDataFor", getSqlCfgHandler("respHandlerPullDataForGet")).Methods("GET")			// Pull all the data for this user ( master is union of all ) - Input URL + User - Output JSON of all notes for page
	mux.HandleFunc("/api/pullDataTopUser", getSqlCfgHandler("respHandlerPullDataTopUserGet")).Methods("GET")	// Pull all the data for painting the top edit page (vititeds, done etc for all URLs, # of notes per url etc)
	mux.HandleFunc("/api/saveOneNote", getSqlCfgHandler("respHandlerSaveOneNoteGet")).Methods("GET")			// Save the data for a single edited/new note (input has pageID)
	mux.HandleFunc("/api/markEditorDone", getSqlCfgHandler("respHandlerMarkEditorDoneGet")).Methods("GET")		// Mark that an editor is done looking at / reviewing a page - they cliked the "done" button for page "D" - has Y/N done flag for "undone"
	mux.HandleFunc("/api/getLogins", getSqlCfgHandler("respHandlerGetLoginsGet")).Methods("GET")				// Get list of users can login as
	mux.HandleFunc("/api/loginAs", getSqlCfgHandler("respHandlerLoginAsGet")).Methods("GET")					// login as a user
	mux.HandleFunc("/api/pullListTopUser", getSqlCfgHandler("respHandlerPullListTopUserGet")).Methods("GET")	// Get list of all articles
	mux.HandleFunc("/api/pullListFor", getSqlCfgHandler("respHandlerPullListForGet")).Methods("GET")			// Get list of articles assiged to user
	mux.HandleFunc("/api/deleteUrl", getSqlCfgHandler("respHandlerDeleteUrlGet")).Methods("GET")				// Delete a URL fro this user - to clean it up, if user is master then delete from list.
	mux.HandleFunc("/api/deleteNote", getSqlCfgHandler("respHandlerDeleteNoteGet")).Methods("GET")				// Delete a note.
	mux.HandleFunc("/api/markToUpdate", getSqlCfgHandler("api/markToUpdate")).Methods("GET")


	mux.HandleFunc("/api/table/{name}", respHandlerTableGet ).Methods("GET")
	mux.HandleFunc("/api/table/{name}", respHandlerTablePut ).Methods("PUT")
	mux.HandleFunc("/api/table/{name}", respHandlerTablePost).Methods("POST")
	mux.HandleFunc("/api/table/{name}", respHandlerTableDel ).Methods("DEL")
	mux.HandleFunc("/api/table/{name}", respHandlerTableHead).Methods("HEAD")

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
