package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

var (
	logDir  string
	logFile *os.File
)

func Init(baseConfigDir string) error {
	logDir = filepath.Join(baseConfigDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}
	ts := time.Now().Format("20060102_150405")
	name := fmt.Sprintf("cronshot_%s.log", ts)
	path := filepath.Join(logDir, name)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	logFile = f
	rotateKeep(5)
	Info("logger initialized")
	return nil
}

func dirEntries() ([]os.DirEntry, error) {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func rotateKeep(max int) {
	entries, err := dirEntries()
	if err != nil {
		return
	}
	type item struct {
		name string
		mod  time.Time
	}
	var items []item
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		items = append(items, item{name: e.Name(), mod: info.ModTime()})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].mod.After(items[j].mod) })
	if len(items) <= max {
		return
	}
	for _, it := range items[max:] {
		_ = os.Remove(filepath.Join(logDir, it.name))
	}
}

func write(prefix, msg string) {
	if logFile == nil {
		return
	}
	line := fmt.Sprintf("%s [%s] %s\n", time.Now().Format("2006-01-02 15:04:05.000"), prefix, msg)
	_, _ = logFile.WriteString(line)
}

func Info(msg string)  { write("INFO", msg) }
func Error(msg string) { write("ERROR", msg) }

func GetLogDir() string { return logDir }
