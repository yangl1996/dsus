package main

import (
    _ "github.com/mattn/go-sqlite3"
    "database/sql"
    "flag"
    "log"
    "os"
    "text/tabwriter"
    "fmt"
)

func db(args []string) {
    l := log.New(os.Stderr, "", 0)
    cmd := flag.NewFlagSet("db", flag.ExitOnError)
    dbPath := cmd.String("db", "", "path to the database")
    listFlag := cmd.Bool("l", false, "list entries")
    insertFlag := cmd.Bool("n", false, "insert a new entry")
    delFlag := cmd.Bool("r", false, "delete an entry")
    modifyFlag := cmd.Bool("m", false, "modify an entry")
    key := cmd.String("k", "", "key of the entry")
    url := cmd.String("u", "", "url of the entry")
    cmd.Parse(args)

    if *dbPath == "" {
        l.Fatalln("missing path to the database")
    }

    if !(*insertFlag || *delFlag || *modifyFlag || *listFlag) {
        l.Fatalln("no action specified")
    }

    numActions := 0
    if *insertFlag {
        numActions += 1
    }
    if *delFlag {
        numActions += 1
    }
    if *modifyFlag {
        numActions += 1
    }
    if numActions > 1 {
        l.Fatalln("insert, delete and modifyFlag should be set at the same time")
    }

    // connect to the database
    db, err := sql.Open("sqlite3", *dbPath)
    if err != nil {
        l.Fatal(err)
    }
    defer db.Close()

    // create the table
    err = initDB(db)
    if err != nil {
        l.Fatal(err)
    }

    if *insertFlag {
        if *key == "" {
            l.Fatalln("missing key")
        }
        if *url == "" {
            l.Fatalln("missing url")
        }
        err = insert(db, key, url)
        if err != nil {
            l.Fatal(err)
        }
    }

    if *delFlag {
        if *key == "" {
            l.Fatalln("missing key")
        }
        err = del(db, key)
        if err != nil {
            l.Fatal(err)
        }
    }

    if *modifyFlag {
        if *key == "" {
            l.Fatalln("missing key")
        }
        if *url == "" {
            l.Fatalln("missing url")
        }
        err = modify(db, key, url)
        if err != nil {
            l.Fatal(err)
        }
    }

    if *listFlag {
        err = list(db)
        if err != nil {
            l.Fatal(err)
        }
    }
}

func initDB(db *sql.DB) error {
    check := `select '' from forwarding limit 1;`
    res := db.QueryRow(check)
    var dest string
    err := res.Scan(&dest)
    // ErrNoRows if table exists but no data
    // Otherwise if the table does not exist at all (or other errors)
    if err == sql.ErrNoRows {
        return nil
    } else if err != nil {
        stmt := `create table forwarding (id integer not null primary key, key text, url text);`
        _, err := db.Exec(stmt)
        return err
    }
    return nil
}

func insert(db *sql.DB, key *string, url *string) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Commit()
    query, err := tx.Prepare("select '' from forwarding where key = ?")
    if err != nil {
        return err
    }
    defer query.Close()
    stmt, err := tx.Prepare("insert into forwarding(key, url) values(?, ?)")
    if err != nil {
        return err
    }
    defer stmt.Close()
    // make sure the entry is not present as of now
    var dest string
    err = query.QueryRow(*key).Scan(&dest)
    if err == nil {
        return nil
    } else if err != sql.ErrNoRows {
        return err
    }
    _, err = stmt.Exec(*key, *url)
    if err != nil {
        return err
    }
    return nil
}

func modify(db *sql.DB, key *string, url *string) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Commit()
    update, err := tx.Prepare("update forwarding set url = ? where key = ?")
    if err != nil {
        return err
    }
    defer update.Close()
    _, err = update.Exec(*url, *key)
    if err != nil {
        return err
    }
    return nil
}

func list(db *sql.DB) error {
    rows, err := db.Query("select key, url from forwarding")
    if err != nil {
        return err
    }
    defer rows.Close()
    w := tabwriter.NewWriter(os.Stdout, 8, 8, 4, ' ', 0)
    defer w.Flush()
    fmt.Fprintln(w, "Key\tURL")
    for rows.Next() {
        var key string
        var url string
        err = rows.Scan(&key, &url)
        if err != nil {
            return err
        }
        fmt.Fprintf(w, "%s\t%s\n", key, url)
    }
    return nil
}

func del(db *sql.DB, key *string) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Commit()
    del, err := tx.Prepare("delete from forwarding where key = ?")
    if err != nil {
        return err
    }
    defer del.Close()
    _, err = del.Exec(*key)
    if err != nil {
        return err
    }
    return nil
}
