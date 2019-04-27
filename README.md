# traceface

API wrapper for **Amazon Rekognition** allows easy extraction of person face's coordinates from video file frames.

### SETUP
(example for default **Go** installation in **Windows**)

1. place project files into

    `%USERPROFILE%\go\src\traceface\`

2. download **AWS SDK**

    `go get -u github.com/aws/aws-sdk-go`


### USAGE

make sure `amazon.ini` in project root directory contains valid API access credentials

example file contents:
```
[Credentials]
AccessKeyID=00000000000000000000
SecretAccessKey=0000000000000000000000000000000000000000
Region=us-west-2
```

also check **S3** bucket existence in the same *Region* as your provided *Credentials*

for this example bucket name is `mybucket` and **S3** object name for new upload is `mykey`

on first run video file uploaded to **Amazon S3** and processed:

    go run traceface.go -s amazon -b mybucket -o mykey -f videofile.mp4

> unique job identifier is stored in `jobs` directory

on second run previously stored job identifier will be used to retrieve processing results:

    go run traceface.go -s amazon

> received data is saved to `spotted` directory

resulting `.ini` file contains sections for every frame where face position change was detected:
```
[f_0_0_0]
Timestamp = 0
FrameN = 0
Left = 0.485185
Top = 0.204861
Width = 0.114815
Height = 0.126389
```

> face bounding box coordinate and dimensions presented as a ratio of the frame size

example of further processing in VirtualDub:

[VirtualDub screenshot](docs/faceblur.png)