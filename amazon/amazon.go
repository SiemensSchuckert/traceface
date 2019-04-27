package amazon

import (
	"context"
	"fmt"
	"os"
	"time"

	"traceface/common"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const configFile = "amazon.ini"

var (
	accessKeyID     = "00000000000000000000"
	secretAccessKey = "0000000000000000000000000000000000000000"
	region          = "us-west-2"
)

var sess *session.Session

func checkError(err error) {

	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case rekognition.ErrCodeInvalidS3ObjectException:
			fmt.Println(rekognition.ErrCodeInvalidS3ObjectException, aerr.Error())
		case rekognition.ErrCodeInvalidParameterException:
			fmt.Println(rekognition.ErrCodeInvalidParameterException, aerr.Error())
		case rekognition.ErrCodeAccessDeniedException:
			fmt.Println(rekognition.ErrCodeAccessDeniedException, aerr.Error())
		case rekognition.ErrCodeInternalServerError:
			fmt.Println(rekognition.ErrCodeInternalServerError, aerr.Error())
		case rekognition.ErrCodeThrottlingException:
			fmt.Println(rekognition.ErrCodeThrottlingException, aerr.Error())
		case rekognition.ErrCodeProvisionedThroughputExceededException:
			fmt.Println(rekognition.ErrCodeProvisionedThroughputExceededException, aerr.Error())
		case rekognition.ErrCodeIdempotentParameterMismatchException:
			fmt.Println(rekognition.ErrCodeIdempotentParameterMismatchException, aerr.Error())
		case rekognition.ErrCodeLimitExceededException:
			fmt.Println(rekognition.ErrCodeLimitExceededException, aerr.Error())
		case rekognition.ErrCodeVideoTooLargeException:
			fmt.Println(rekognition.ErrCodeVideoTooLargeException, aerr.Error())
		case rekognition.ErrCodeInvalidPaginationTokenException:
			fmt.Println(rekognition.ErrCodeInvalidPaginationTokenException, aerr.Error())
		case rekognition.ErrCodeResourceNotFoundException:
			fmt.Println(rekognition.ErrCodeResourceNotFoundException, aerr.Error())
		default:
			fmt.Println(aerr.Error())
		}
	} else {
		fmt.Println(err.Error())
	}
}

func checkMeta(vm *rekognition.VideoMetadata) {
	if vm.DurationMillis == nil || *vm.DurationMillis < 1000 {
		fmt.Println("VideoMetadata DurationMillis parse error...")
		os.Exit(1)
	}
	if vm.FrameRate == nil || *vm.FrameRate < 1.0 {
		fmt.Println("VideoMetadata FrameRate parse error...")
		os.Exit(1)
	}
	if vm.FrameWidth == nil || *vm.FrameWidth < 1 {
		fmt.Println("VideoMetadata FrameWidth parse error...")
		os.Exit(1)
	}
	if vm.FrameHeight == nil || *vm.FrameHeight < 1 {
		fmt.Println("VideoMetadata FrameHeight parse error...")
		os.Exit(1)
	}
}

func outBox(persons []*rekognition.PersonDetection, vm *rekognition.VideoMetadata) {
	var f *os.File
	var p_index int64
	var str string
	var err error

	for num, person := range persons {
		if person.Person != nil && person.Person.Index != nil && person.Person.Face != nil {
			//fmt.Println(person.Person.Face)

			if p_index != *person.Person.Index && f != nil {
				f.Sync()
				err = f.Close()
				if err != nil {
					fmt.Fprintf(os.Stderr, "File close error, %v\n", err)
					os.Exit(1)
				}
				f = nil
			}
			if f == nil {
				str = common.JobFile(*person.Person.Index)
				f, err = os.OpenFile(str, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Cannot create file %q, %v\n", str, err)
					os.Exit(1)
				}
			}
			p_index = *person.Person.Index

			common.JobSetSection(num, *person.Timestamp, p_index)
			common.JobSetTimestamp(*person.Timestamp)
			common.JobSetFrame(int64((float64(*person.Timestamp) / 1000.0) * *vm.FrameRate))
			common.JobSetBoxLeft(float64(*person.Person.Face.BoundingBox.Left))
			common.JobSetBoxTop(float64(*person.Person.Face.BoundingBox.Top))
			common.JobSetBoxWidth(float64(*person.Person.Face.BoundingBox.Width))
			common.JobSetBoxHeight(float64(*person.Person.Face.BoundingBox.Height))

			_, err := f.WriteString(common.JobGetResult())
			if err != nil {
				fmt.Fprintf(os.Stderr, "Cannot write to file %q, %v\n", str, err)
				os.Exit(1)
			}

		}
	}
	if f != nil {
		f.Sync()
		f.Close()
	}
}

func getTrack(job string) {

	pager := func(page *rekognition.GetPersonTrackingOutput, lastPage bool) bool {
		if page.JobStatus != nil {
			switch *page.JobStatus {
			case "FAILED":
				fmt.Println(*page.StatusMessage)
				os.Exit(1)
			case "IN_PROGRESS":
				fmt.Println("WORK IN PROGRESS! Please try again later.")
				os.Exit(0)
			case "SUCCEEDED":
				if page.Persons != nil && page.VideoMetadata != nil {
					checkMeta(page.VideoMetadata)
					outBox(page.Persons, page.VideoMetadata)
				}
			default:
				fmt.Println("Unknown job status...")
				os.Exit(1)
			}
		}

		return true
	}

	svc := rekognition.New(sess)
	input := &rekognition.GetPersonTrackingInput{
		JobId: aws.String(job),
		//MaxResults: aws.Int64(500),
	}

	err := svc.GetPersonTrackingPages(input, pager)

	if err != nil {
		checkError(err)
		os.Exit(1)
	}

	common.JobDone()
}

func startTrack(bucket, key string) {
	svc := rekognition.New(sess)
	input := &rekognition.StartPersonTrackingInput{
		Video: &rekognition.Video{
			S3Object: &rekognition.S3Object{
				Bucket: aws.String(bucket),
				Name:   aws.String(key),
			},
		},
	}

	result, err := svc.StartPersonTracking(input)

	if err != nil {
		checkError(err)
		os.Exit(1)
	}

	if result.JobId != nil {
		common.JobStore(*result.JobId)
		fmt.Println("Job started:")
		fmt.Println(common.JobName)
	} else {
		fmt.Println("Cannot start job...")
		os.Exit(1)
	}
}

func startUpload(fname, bucket, key string, timeout time.Duration) {

	f, err := os.OpenFile(fname, os.O_RDONLY, 0444)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open file %q, %v\n", fname, err)
		os.Exit(1)
	}
	fmt.Println("Upload in progress...")
	uploader := s3manager.NewUploader(sess)

	ctx := context.Background()
	var cancelFn func()
	if timeout > 0 {
		ctx, cancelFn = context.WithTimeout(ctx, timeout)
		defer cancelFn()
	}

	_, err = uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   f,
	})

	if f != nil {
		f.Close()
	}

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
			fmt.Fprintf(os.Stderr, "Upload stop: timeout, %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Upload failed, %v\n", err)
		}
		os.Exit(1)
	}
}

func StartSession(job, infile, bucket, key string, timeout time.Duration) {

	fullConf := common.ParseConfig(configFile)
	for _, line := range fullConf {

		switch line.Name {
		case "AccessKeyID":
			accessKeyID = line.Value

		case "SecretAccessKey":
			secretAccessKey = line.Value

		case "Region":
			region = line.Value
		}
	}

	sess = session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	}))

	job = common.Trims(job)
	if job != "" {
		getTrack(job)
	} else {
		startUpload(infile, bucket, key, timeout)
		startTrack(bucket, key)
	}

}
