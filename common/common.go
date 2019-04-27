package common

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"
)

const (
	OutDir = "spotted"
	JobDir = "jobs"
)

type cfgField struct {
	Name  string
	Value string
}

type conf []cfgField

var TS string
var JobName string
var resultSection string
var Trims = strings.TrimSpace

func TrimQ(ins string) string {
	return strings.Trim(ins, "\\'\"")
}

var tr = &http.Transport{
	MaxIdleConns:       10,
	IdleConnTimeout:    10 * time.Second,
	DisableCompression: true,
}

var htclient = &http.Client{Transport: tr}

func HTTPSend(request *http.Request) []byte {

	resp, err := htclient.Do(request)
	if err != nil {
		fmt.Fprintf(os.Stderr, "HTTP request failed with error: %v\n", err)
		os.Exit(1)
	}

	if resp.ContentLength < 1 {
		fmt.Fprintf(os.Stderr, "HTTP bad response length\n")
		os.Exit(1)
	}

	reqResult := make([]byte, resp.ContentLength)
	num, _ := resp.Body.Read(reqResult)

	if resp.Body != nil {
		resp.Body.Close()
	}

	if num < 1 {
		fmt.Fprintf(os.Stderr, "HTTP failed to read response\n")
		os.Exit(1)
	}

	return reqResult
}

func RemoveSpaces(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

func ParseConfig(src string) conf {
	var num int
	var fullConf conf
	cfg := make([]byte, 16384)

	f, _ := os.OpenFile(src, os.O_RDONLY, 0444)

	if f != nil {
		num, _ = f.Read(cfg)
	}

	if f != nil {
		f.Close()
	}

	if num < 10 {
		return fullConf
	}

	res := strings.Split(string(cfg[:num]), "\n")

	for _, field := range res {
		fpair := strings.SplitN(RemoveSpaces(field), "=", 2)
		if len(fpair) == 2 {
			ff := cfgField{fpair[0], fpair[1]}
			fullConf = append(fullConf, ff)
		}
	}

	return fullConf
}

func JobRemove() {
	_ = os.Remove(fmt.Sprintf("%s%s%s", JobDir, "/", "current.txt"))
	fmt.Println("Unfinished job removed!")
	os.Exit(0)
}

func JobDone() {
	_ = os.Remove(fmt.Sprintf("%s%s%s", JobDir, "/", "current.txt"))
	fmt.Println("Job done:")
	fmt.Println(JobName)
}

func JobStore(jobID string) {
	str := fmt.Sprintf("%s%s%s%s", JobDir, "/", JobName, ".txt")
	f, err := os.OpenFile(str, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot create file %q, %v\n", str, err)
		os.Exit(1)
	}
	_, err = f.WriteString(jobID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot write to file %q, %v\n", str, err)
		os.Exit(1)
	}
	if f != nil {
		f.Close()
	}

	str = fmt.Sprintf("%s%s", JobDir, "/current.txt")
	f, err = os.OpenFile(str, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot create file, %v\n", err)
		os.Exit(1)
	}

	str = fmt.Sprintf("%s", JobName)
	_, err = f.WriteString(str)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot write to file, %v\n", err)
		os.Exit(1)
	}
	if f != nil {
		f.Close()
	}
}

func JobFile(index int64) string {
	return fmt.Sprintf("%s%s%s%s%d%s", OutDir, "/", JobName, "_", index, ".ini")
}

func JobGetResult() string {
	return resultSection
}

func JobSetSection(num int, s_timestamp, s_index int64) {
	resultSection = fmt.Sprintf("%s%d%s%d%s%d%s", "[f_", num, "_", s_timestamp, "_", s_index, "]\n")
}

func JobSetTimestamp(s_timestamp int64) {
	resultSection = fmt.Sprintf("%s%s%d%s", resultSection, "Timestamp=", s_timestamp, "\n")
}

func JobSetFrame(s_frame int64) {
	resultSection = fmt.Sprintf("%s%s%d%s", resultSection, "FrameN=", s_frame, "\n")
}

func JobSetBoxLeft(s_left float64) {
	resultSection = fmt.Sprintf("%s%s%f%s", resultSection, "Left=", s_left, "\n")
}

func JobSetBoxTop(s_top float64) {
	resultSection = fmt.Sprintf("%s%s%f%s", resultSection, "Top=", s_top, "\n")
}

func JobSetBoxWidth(s_width float64) {
	resultSection = fmt.Sprintf("%s%s%f%s", resultSection, "Width=", s_width, "\n")
}

func JobSetBoxHeight(s_height float64) {
	resultSection = fmt.Sprintf("%s%s%f%s", resultSection, "Height=", s_height, "\n")
}
