package script

import (
	"github.com/sentrycloud/sentry/pkg/agent/reporter"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Script struct {
	path       string
	scriptType string // python, bash, ...
	interval   int
}

type FileScanner struct {
	fileList []string
}

func StartScriptScheduler(scriptDirs string, scriptType string, report *reporter.Reporter) {
	var fileScanner = FileScanner{}
	scripts := fileScanner.GetAllFiles(scriptDirs)
	for _, script := range scripts {
		interval, err := getScheduleInterval(scriptType, script)
		if err != nil {
			continue
		}
		s := &Script{path: script, scriptType: scriptType, interval: interval}
		scheduler := NewScheduler(s, report)
		scheduler.Start()
	}
}

func getScheduleInterval(scriptType string, script string) (int, error) {
	out, err := RunCommand(10, scriptType, script, "sentry_time")
	if err != nil {
		return 0, err
	}

	interval, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		newlog.Error("sentry_time is not an integer: %v", err)
		return 0, err
	}

	return interval, nil
}

func (s *FileScanner) GetAllFiles(root string) []string {
	err := filepath.WalkDir(root, s.filterFile)
	if err != nil {
		newlog.Error("WalkDir failed: %v", err)
	}
	return s.fileList
}

func (s *FileScanner) filterFile(path string, d os.DirEntry, err error) error {
	if err != nil {
		newlog.Error("error access path %s: %v", path, err)
		return nil
	}

	if !d.IsDir() {
		s.fileList = append(s.fileList, path)
	}

	return nil
}
