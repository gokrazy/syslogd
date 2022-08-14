package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestColdLogFileNames(t *testing.T) {
	srv := server{
		dir:   t.TempDir(),
		files: make(map[fileKey]*openFile),
	}
	for _, rel := range []string{
		"dr/2022-08-10.log",
		"dr/2022-08-11.log",
		"dr/2022-08-12.log",
		"dr/2022-08-13.log",
		"router7/2022-08-10.log",
	} {
		fn := filepath.Join(srv.dir, rel)
		if err := os.MkdirAll(filepath.Dir(fn), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fn, nil, 0644); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Date(2022, time.August, 13, 16, 20, 0, 0, time.Local)
	cold, err := srv.coldLogFileNames(now)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		filepath.Join(srv.dir, "dr", "2022-08-10.log"),
		filepath.Join(srv.dir, "dr", "2022-08-11.log"),
		// 2022-08-12.log might still be in use (old messages)
		// 2022-08-13.log is definitely in use
		filepath.Join(srv.dir, "router7", "2022-08-10.log"),
	}
	if diff := cmp.Diff(want, cold); diff != "" {
		t.Errorf("coldLogFileNames(): unexpected diff (-want +got):\n%s", diff)
	}
}

func TestColdLogFileNamesSingle(t *testing.T) {
	srv := server{
		dir:   t.TempDir(),
		files: make(map[fileKey]*openFile),
	}
	for _, rel := range []string{
		"dr/2022-08-13.log",
	} {
		fn := filepath.Join(srv.dir, rel)
		if err := os.MkdirAll(filepath.Dir(fn), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fn, nil, 0644); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Date(2022, time.August, 13, 16, 20, 0, 0, time.Local)
	cold, err := srv.coldLogFileNames(now)
	if err != nil {
		t.Fatal(err)
	}
	var want []string
	if diff := cmp.Diff(want, cold); diff != "" {
		t.Errorf("coldLogFileNames(): unexpected diff (-want +got):\n%s", diff)
	}
}

func TestCompressFile(t *testing.T) {
	const contents = "hello syslog"
	fn := filepath.Join(t.TempDir(), "2022-08-10.log")
	if err := os.WriteFile(fn, []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
	if err := compressFile(fn); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	gunzip := exec.Command("zstdcat", fn+".zst")
	gunzip.Stdout = &buf
	gunzip.Stderr = os.Stderr
	if err := gunzip.Run(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(contents, buf.String()); diff != "" {
		t.Fatalf("compressFile: unexpected diff (-want +got):\n%s", diff)
	}
}
