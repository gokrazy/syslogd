package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
)

func grog(ctx context.Context) error {
	var (
		hostname = flag.String("hostname",
			"dr",
			"hostname to grep the log for")

		base = flag.String("web_base",
			"http://router7:8514",
			"base URL of gokr-syslogweb service to query")

		grepRange = flag.String("range",
			"todayyesterday",
			"syslog range to grep; one of todayyesterday or all")
	)
	flag.Parse()

	if flag.NArg() != 1 {
		return fmt.Errorf("syntax: grog [--hostname=<host>] <grep pattern>")
	}
	pattern := flag.Arg(0)

	u, err := url.Parse(*base)
	if err != nil {
		return err
	}
	u.Path = "/grep/" + *hostname
	q := u.Query()
	q.Set("q", pattern)
	q.Set("range", *grepRange)
	u.RawQuery = q.Encode()
	log.Printf("Grepping syslog via HTTP: %s", u)

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP response code: got %v, want %v", resp.Status, http.StatusOK)
	}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "rfc3339=") {
			if idx := strings.IndexByte(line, ' '); idx > -1 {
				line = line[idx+1:]
			}
		}
		os.Stdout.WriteString(line)
		os.Stdout.Write([]byte{'\n'})
	}
	return scanner.Err()
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := grog(ctx); err != nil {
		log.Fatal(err)
	}
}
