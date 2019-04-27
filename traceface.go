package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"traceface/amazon"
	"traceface/common"
	//"traceface/google"
)

func getJob(jobName string) string {
	var num int
	job := make([]byte, 128)

	if jobName == "" {
		str := fmt.Sprintf("%s%s", common.JobDir, "/current.txt")
		f, _ := os.OpenFile(str, os.O_RDONLY, 0444)

		if f != nil {
			num, _ = f.Read(job)
		}

		if f != nil {
			f.Close()
		}

		if num < 3 {
			return ""
		}

		common.JobName = string(job[:num])
	} else {
		common.JobName = jobName
	}

	str := fmt.Sprintf("%s%s%s%s", common.JobDir, "/", common.JobName, ".txt")
	f, _ := os.OpenFile(str, os.O_RDONLY, 0444)

	num = 0

	if f != nil {
		num, _ = f.Read(job)
	}

	if f != nil {
		f.Close()
	}

	if num > 0 {
		return string(job[:num])
	}
	return ""
}

func main() {
	var rservice, job, bucket, key, infile string
	var rem bool
	var timeout time.Duration

	flag.StringVar(&rservice, "s", "", "Service to use")
	flag.StringVar(&job, "j", "", "Job name")
	flag.BoolVar(&rem, "r", false, "Remove pending job")
	flag.StringVar(&bucket, "b", "", "Bucket name")
	flag.StringVar(&key, "o", "", "Object key name")
	flag.DurationVar(&timeout, "d", 0, "Upload timeout")
	flag.StringVar(&infile, "f", "", "Input file")
	flag.Parse()

	common.TS = fmt.Sprintf("%X", time.Now().Unix())
	err := os.MkdirAll(common.JobDir, os.ModePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot create directory %q, %v\n", common.JobDir, err)
		os.Exit(1)
	}
	err = os.MkdirAll(common.OutDir, os.ModePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot create directory %q, %v\n", common.OutDir, err)
		os.Exit(1)
	}

	if rem {
		common.JobRemove()
	}

	job = getJob(job)

	if job == "" {
		if infile == "" {
			fmt.Fprintf(os.Stderr, "Input file not specified ( -f filename )\n")
			os.Exit(2)
		}
		_, common.JobName = filepath.Split(infile)
		common.JobName = common.RemoveSpaces(common.JobName)
		ife := filepath.Ext(common.JobName)
		common.JobName = common.JobName[:len(common.JobName)-len(ife)]
		common.JobName = fmt.Sprintf("%s%s%s", common.JobName, "_", common.TS)
	}

	switch rservice {
	case "amazon":
		amazon.StartSession(job, infile, bucket, key, timeout)
	//case "google":
		//google.StartSession(job, infile, timeout)
	default:
		fmt.Fprintf(os.Stderr, "Service not selected ( -s servicename )\n")
		os.Exit(2)
	}

}
