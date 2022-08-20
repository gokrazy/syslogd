package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestToDeleteLogFileNames(t *testing.T) {
	srv := server{
		dir:   t.TempDir(),
		files: make(map[fileKey]*openFile),
	}
	for _, rel := range []string{
		"dr/2022-08-10.log",
		"dr/2022-08-11.log",
		"dr/2022-08-12.log",
		"dr/2022-08-13.log",
		"dr/2022-08-14.log",
		"dr/2022-08-15.log",
		"dr/2022-08-16.log",
		"dr/2022-08-17.log",
		"dr/2022-08-18.log",
		"router7/2022-08-10.log",
		// intentional gap
		"router7/2022-08-18.log",
	} {
		fn := filepath.Join(srv.dir, rel)
		if err := os.MkdirAll(filepath.Dir(fn), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fn, nil, 0644); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Date(2022, time.August, 18, 16, 20, 0, 0, time.Local)
	cold, err := srv.toDeleteLogFileNames(now)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		filepath.Join(srv.dir, "dr", "2022-08-10.log"),
		filepath.Join(srv.dir, "router7", "2022-08-10.log"),
	}
	if diff := cmp.Diff(want, cold); diff != "" {
		t.Errorf("toDeleteLogFileNames(): unexpected diff (-want +got):\n%s", diff)
	}
}
