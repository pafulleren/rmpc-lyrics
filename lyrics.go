// go run ./lyrics/example "Daft Punk" "Get Lucky"
package main

import (
	"context"
	"fmt"
	"github.com/AvengeMedia/dankgo/lyrics"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

func main() {
	if os.Getenv("HAS_LRC") == "true" {
		os.Exit(0)
	}

	pid := os.Getenv("PID")
	lrcFile := os.Getenv("LRC_FILE")
	os.MkdirAll(path.Dir(lrcFile), 0755)

	var hasAlbum bool
	if os.Getenv("ALBUM") != "" {
		hasAlbum = true
	} else {
		hasAlbum = false
	}
	artist := os.Getenv("ARTIST")
	title := os.Getenv("TITLE")
	if artist == "" || title == "" {
		os.Exit(0)
	}
	var album string
	if hasAlbum {
		album = os.Getenv("ALBUM")
	}
	req := lyrics.Request{
		Artist: artist,
		Title:  title,
	}
	if hasAlbum {
		req.Album = album
	}

	client := lyrics.New(lyrics.Options{UserAgent: "dankgo-lyrics-example (+https://github.com/AvengeMedia/dankgo)"})
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := client.Lookup(ctx, req)
	if err != nil {
		cmd := exec.Command("rmpc", "remote", "--pid", pid, "status", "Lyrics download failed", "--level", "error")
		cmd.Run()
		os.Exit(1)
	}
	if !result.Found {
		cmd := exec.Command("rmpc", "remote", "--pid", pid, "status", "Lyrics not found.", "--level", "warn")
		cmd.Run()
		os.Exit(0)
		return
	}

	if len(result.Synced) == 0 {
		cmd := exec.Command("rmpc", "remote", "--pid", pid, "status", "No synced lyrics.")
		cmd.Run()
		os.Exit(0)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "[ti:%s]\n[ar:%s]\n", title, artist)
	if hasAlbum {
		fmt.Fprintf(&sb, "[al:%s]\n\n", album)
	} else {
		fmt.Fprintf(&sb, "\n")
	}
	for _, line := range result.Synced {
		fmt.Fprintf(&sb, "%s%s\n", formatSeconds(line.Start), line.Text)
	}
	file, err := os.Create(lrcFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	_, err = file.WriteString(sb.String())
	if err != nil {
		panic(err)
	}
	cmd := exec.Command("rmpc", "remote", "--pid", pid, "indexlrc", "--file", lrcFile)
	cmd2 := exec.Command("rmpc", "remote", "--pid", pid, "status", "Lyrics downloaded.", "--level", "info")
	cmd.Run()
	cmd2.Run()
}

func formatSeconds(sec lyrics.Seconds) string {
	duration := time.Duration(sec * lyrics.Seconds(time.Second))

	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) % 60
	milliseconds := (duration.Milliseconds()) % 1000

	return fmt.Sprintf("[%02d:%02d.%03d]", minutes, seconds, milliseconds)
}
