// Binary gokr-syslogweb is a web server that provides access to the syslog
// files that gokr-syslogd writes.
package main

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/klauspost/compress/zstd"
)

type errorHTTPHandler func(http.ResponseWriter, *http.Request) error

func middleware(h errorHTTPHandler) http.Handler {
	// Could extend this later
	return handleError(h)
}

//go:embed *.html.tmpl
var templateFiles embed.FS

var indexTmpl = template.Must(template.New("index.html.tmpl").Funcs(template.FuncMap{
	"formatTime": func(t time.Time) string {
		return t.Format("2006-01-02 15:04:05")
	},
}).ParseFS(templateFiles, "index.html.tmpl"))

const basenameFormat = "2006-01-02.log"

func syslogweb() error {
	// TODO: listen on (all?) gokrazy private IPs by default
	var (
		syslogdDir = flag.String("syslogd_dir",
			"/perm/syslogd",
			"directory to which to serve syslogs from")

		listenAddrs = flag.String("listen",
			"localhost:8514", // 514 is syslog, 80 is web
			"comma-separated list of [host]:port pairs to listen on")
	)

	flag.Parse()

	mux := http.NewServeMux()

	mux.Handle("/grep/", middleware(func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()

		host := strings.TrimPrefix(r.URL.Path, "/grep/")
		if host == "" {
			return httpError(http.StatusNotFound, fmt.Errorf("not found"))
		}

		q := r.FormValue("q")
		if q == "" {
			return httpError(http.StatusBadRequest, fmt.Errorf("empty pattern (q= parameter)"))
		}
		re, err := regexp.Compile(q)
		if err != nil {
			return httpError(http.StatusBadRequest, fmt.Errorf("invalid Go regexp: %q: %v", q, err))
		}

		timeRange := r.FormValue("range")
		if timeRange == "" {
			timeRange = "todayyesterday"
		}
		if timeRange != "todayyesterday" &&
			timeRange != "all" {
			return httpError(http.StatusBadRequest, fmt.Errorf("invalid range= parameter (expected one of todayyesterday or all)"))
		}

		fis, err := os.ReadDir(*syslogdDir)
		if err != nil {
			return err
		}
		hosts := make(map[string]bool)
		for _, fi := range fis {
			hosts[fi.Name()] = true
		}
		if !hosts[host] {
			return httpError(http.StatusNotFound, fmt.Errorf("host %q not found", host))
		}

		now := time.Now()
		var files []string
		if timeRange == "all" {
			fis, err := os.ReadDir(filepath.Join(*syslogdDir, host))
			if err != nil {
				return err
			}
			for _, fi := range fis {
				files = append(files, fi.Name())
			}
		} else {
			yesterday := now.Add(-24 * time.Hour).Format(basenameFormat)
			today := now.Format(basenameFormat)
			files = []string{
				yesterday,
				today,
			}
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		for _, fn := range files {
			f, err := os.Open(filepath.Join(*syslogdDir, host, fn))
			if err != nil {
				return err
			}
			defer f.Close()
			rd := io.Reader(f)
			if strings.HasSuffix(fn, ".zst") {
				dec, err := zstd.NewReader(f)
				if err != nil {
					return err
				}
				defer dec.Close()
				rd = dec
			}
			scanner := bufio.NewScanner(rd)
			for scanner.Scan() {
				if err := ctx.Err(); err != nil {
					return err
				}
				line := scanner.Bytes()
				if !re.Match(line) {
					continue
				}
				if _, err := w.Write(append(line, '\n')); err != nil {
					return err
				}
			}
			if err := scanner.Err(); err != nil {
				return err
			}
			if err := f.Close(); err != nil {
				return err
			}
		}

		return nil
	}))

	mux.Handle("/", middleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.URL.Path != "/" {
			return httpError(http.StatusNotFound, fmt.Errorf("not found"))
		}

		fis, err := os.ReadDir(*syslogdDir)
		if err != nil {
			return err
		}
		hosts := make([]string, 0, len(fis))
		for _, fi := range fis {
			hosts = append(hosts, fi.Name())
		}

		tmplData := struct {
			Hosts []string
		}{
			Hosts: hosts,
		}
		var tmplBuf bytes.Buffer
		if err := indexTmpl.Execute(&tmplBuf, tmplData); err != nil {
			return err
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, &tmplBuf)
		return err
	}))

	addrs := strings.Split(*listenAddrs, ",")
	log.Printf("listening on %q", addrs)
	return multiListen(context.Background(), mux, addrs)
}

func main() {
	if err := syslogweb(); err != nil {
		log.Fatal(err)
	}
}
