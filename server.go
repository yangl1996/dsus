package main

import (
    "flag"
    "log"
    "os"
)

func server(args []string) {
    l := log.New(os.Stderr, "", 0)
    cmd := flag.NewFlagSet("server", flag.ExitOnError)
    cert := cmd.String("cert", "", "path to the SSL certificate")
    cmd.Parse(args)

    if *cert == "" {
        l.Fatalln("missing path to the certificate")
    }


}
