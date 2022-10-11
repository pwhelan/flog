package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Generate generates the logs with given options
func Generate(option *Option) error {
	var (
		splitCount = 1
		created    = time.Now()

		interval time.Duration
		delay    time.Duration
	)

	if option.Sleep > 0 {
		interval = option.Sleep
	}

	logFileName := option.Output
	writer, err := NewWriter(option.Type, logFileName)
	flogWriter, err := NewWriter("log", "/var/log/flog.tlog")
	if err != nil {
		return err
	}

	if option.Forever {
		_, _ = flogWriter.Write([]byte("flog started at " + created.Format(time.RFC3339)))
		for counter := 0; counter < option.Number/option.Rate; counter++ {
			start := time.Now()
			log := ""
			for i := 0; i < option.Rate; i++ {
				log = `{"level":"warn","ts":"2021-05-18T09:58:24.423Z","msg":"client Profile already loaded doniv","hostname":"ovk8mp-perf-060-12-nq9yk-57967d89b7-h4vdx","pid":"%s","src":"profile@v0.6.0-rc.4/client.go:542","stacktrace":"go.uber.org/zap.Stack\\n\\t/go/pkg/mod/go.uber.org/zap@v1.4.1/field.go:209\\ngo.uber.org/zap.(*Logger).check\\n\\t/go/pkg/mod/go.uber.org/zap@v1.4.1/logger.go:273\\ngo.uber.org/zap.(*Logger).Warn\\n\\t/go/pkg/mod/go.uber.org/zap@v1.4.1/logger.go:168\\nvisa.com/ovn/commons/logging.(*Logger).Warn\\n\\t/go/pkg/mod/visa.com/ovn/commons@v0.6.0-rc.2/logging/logger.go:92\\nvisa.com/ovn/components/profile.(*Client).loadProfile\\n\\t/go/pkg/mod/visa.com/ovn/components/profile@v0.6.0-rc.4/client.go:542\\nvisa.com/ovn/components/profile.(*Client).openProfilesToLoad\\n\\t/go/pkg/mod/visa.com/ovn/components/profile@v0.6.0-rc.4/client.go:203\\nvisa.com/ovn/components/profile.(*Client).selfHeal\\n\\t/go/pkg/mod/visa.com/ovn/components/profile@v0.6.0-rc.4/client.go:358\\nvisa.com/ovn/components/profile.(*Client).refreshHealth\\n\\t/go/pkg/mod/visa.com/ovn/components/profile@v0.6.0-rc.4/client.go:326\\nvisa.com/ovn/components/profile.(*Client).Readiness\\n\\t/go/pkg/mod/visa.com/ovn/components/profile@v0.6.0-rc.4/client.go:316\\nvisa.com/ovn/commons/health.invokeRoutine\\n\\t/go/pkg/mod/visa.com/ovn/commons@v0.6.0-rc.2/health/healthmonitor.go:138}\n"}`
				recordId := fmt.Sprintf("%s-%d", start.Format(time.RFC3339), i)
				fmtLog := fmt.Sprintf(log, recordId)
				_, _ = writer.Write([]byte(fmtLog + "\n"))
				created = created.Add(interval)
			}
			elapsed := time.Since(start)
			_, _ = flogWriter.Write([]byte(time.Now().String() + " wrote " + strconv.Itoa(option.Rate) + " logs in " + elapsed.String() + " iteration: " + strconv.Itoa(counter) + "\n"))
			time.Sleep(time.Second - elapsed)
		}
	} else {

		if option.Number > 0 {
			// Generates the logs until the certain number of lines is reached
			for line := 0; line < option.Number; line++ {
				time.Sleep(delay)
				log := NewLog(option.Format, created, option.Bytes)
				_, _ = writer.Write([]byte(log + "\n"))

				if (option.Type != "stdout") && (option.SplitBy > 0) && (line > option.SplitBy*splitCount) {
					_ = writer.Close()
					fmt.Println(logFileName, "is created.")

					logFileName = NewSplitFileName(option.Output, splitCount)
					writer, _ = NewWriter(option.Type, logFileName)

					splitCount++
				}
				created = created.Add(interval)
			}
		}
	}

	_, _ = flogWriter.Write([]byte("flog finished at " + time.Now().Format(time.RFC3339)))
	_, _ = flogWriter.Write([]byte("will wait for " + option.Sleep.String() + " before exiting"))
	_ = flogWriter.Close()

	time.Sleep(option.Sleep)
	if option.Type != "stdout" {
		_ = writer.Close()
		fmt.Println(logFileName, "is created.")
	}

	return nil
}

// NewWriter returns a closeable writer corresponding to given log type
func NewWriter(logType string, logFileName string) (io.WriteCloser, error) {
	switch logType {
	case "stdout":
		return os.Stdout, nil
	case "log":
		logFile, err := os.Create(logFileName)
		if err != nil {
			return nil, err
		}
		return logFile, nil
	case "gz":
		logFile, err := os.Create(logFileName)
		if err != nil {
			return nil, err
		}
		return gzip.NewWriter(logFile), nil
	default:
		return nil, nil
	}
}

// NewLog creates a log for given format
func NewLog(format string, t time.Time, length int) string {
	switch format {
	case "apache_common":
		return NewApacheCommonLog(t)
	case "apache_combined":
		return NewApacheCombinedLog(t)
	case "apache_error":
		return NewApacheErrorLog(t, length)
	case "rfc3164":
		return NewRFC3164Log(t, length)
	case "rfc5424":
		return NewRFC5424Log(t, length)
	case "common_log":
		return NewCommonLogFormat(t)
	case "json":
		return NewJSONLogFormat(t)
	case "spring_boot":
		return NewSpringBootLogFormat(t, length)
	default:
		return ""
	}
}

// NewSplitFileName creates a new file path with split count
func NewSplitFileName(path string, count int) string {
	logFileNameExt := filepath.Ext(path)
	pathWithoutExt := strings.TrimSuffix(path, logFileNameExt)
	return pathWithoutExt + strconv.Itoa(count) + logFileNameExt
}
