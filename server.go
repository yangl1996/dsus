package main

import (
    "flag"
    "log"
    "os"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "net/http"
    "strings"
)

func server(args []string) {
    l := log.New(os.Stderr, "", 0)
    cmd := flag.NewFlagSet("server", flag.ExitOnError)
    sslFlag := cmd.Bool("ssl", false, "enable HTTPS")
    certPath := cmd.String("cert", "", "path to the SSL certificate")
    privkeyPath := cmd.String("privkey", "", "path to the SSL private key")
    dbPath := cmd.String("db", "dsus.sqlite", "path to the database")
    cmd.Parse(args)

    if *sslFlag {
        if *certPath == "" {
            l.Fatalln("missing path to the certificate")
        }
        if *privkeyPath == "" {
            l.Fatalln("missing path to the private key")
        }
    }

    if *dbPath == "" {
        l.Fatalln("missing path to the database")
    }

    db, err := sql.Open("sqlite3", *dbPath)
    if err != nil {
        l.Fatal(err)
    }
    defer db.Close()

    err = initDB(db)
    if err != nil {
        l.Fatal(err)
    }
    h, err := NewUrlExpandHandler(db)
    if err != nil {
        l.Fatal(err)
    }
    if !*sslFlag {
	    log.Println("Server starting (plain-text HTTP)")
	    err = http.ListenAndServe(":80", h)
	    if err != nil {
		    l.Fatal(err)
	    }
    } else {
	    log.Println("Server started (SSL)")
	    err = http.ListenAndServeTLS(":443", *certPath, *privkeyPath, h)
	    if err != nil {
		    l.Fatal(err)
	    }
    }
}

type urlExpandHandler struct {
    db *sql.DB
    query *sql.Stmt
}

func NewUrlExpandHandler(db *sql.DB) (http.Handler, error) {
    h := new(urlExpandHandler)
    stmt, err := db.Prepare("select url from forwarding where key = ?")
    if err != nil {
        return h, err
    }
    h.db = db
    h.query = stmt
    return h, nil
}

func (h *urlExpandHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path
    path = strings.Trim(path, "/")
    if r.Method != "GET" {
        http.Error(w, "Method Not Allowed", 405)
        log.Printf("Method Not Allowed %v %v\n", r.Method, path)
        return
    }
    var expanded string
    err := h.query.QueryRow(path).Scan(&expanded)
    if err == sql.ErrNoRows {
        http.NotFound(w, r)
        log.Printf("Not Found %v\n", path)
        return
    } else if err != nil {
        http.Error(w, "Service Unavailable", 503)
        log.Printf("Service Unavailable %v: %v\n", path, err)
        return
    }

    http.Redirect(w, r, expanded, 307)
    log.Printf("Forwarded %v -> %v\n", path, expanded)
    return
}
