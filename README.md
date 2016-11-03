# http2s3
Process form POSTs, pushing files to S3 and notifying over HipChat (or whatever)

h23 is designed to run unattended using EC2 and Elastic Beanstalk in conjunction 
with IAM roles assigned to the instances. I've omitted the Beanstalk configs, but
otherwise this is a snap to get out and scale with your upload load.

Also supported is a minimal system-level config that points to an Etcd cluster.
The code for that was removed prior to this release as it needs some razorblades
taken out of the sandbox, but it works great for environments configured exactly 
like mine. :)

## CLI

There are lots of command line options for you to not use.
```go
$ ./http2s3 --help
Usage of ./http2s3:
  -awsaccesskey string
    	AWS access key
  -awsregion string
    	AWS region
  -awssecretkey string
    	AWS secret key
  -bind string
    	Address to bind on. If this value has a colon, as in ":8000" or
		"127.0.0.1:9001", it will be treated as a TCP address. If it
		begins with a "/" or a ".", it will be treated as a path to a
		UNIX socket. If it begins with the string "fd@", as in "fd@3",
		it will be treated as a file descriptor (useful for use with
		systemd, for instance). If it begins with the string "einhorn@",
		as in "einhorn@0", the corresponding einhorn socket will be
		used. If an option is not explicitly passed, the implementation
		will automatically select among "einhorn@0" (Einhorn), "fd@3"
		(systemd), and ":8000" (fallback) based on its environment. (default ":8000")
  -bucket string
    	S3 Bucket to put files in
  -configfolder string
    	Where the config files live (default "configs/")
  -debug
    	Enable Debug output
  -from string
    	HipChat username to send from
  -getredirect string
    	On GET (vs. POST) where to redirect the misguided to
  -room string
    	HipChat room to send to
  -temproot string
    	Where to create temp folders and files (default "/tmp/")
  -token string
    	HipChat token to use
```

## Configs

You may have noticed there's a default place config files live? Toss a .json file like 
the one below in that folder (relative to wherever you're launching from) and magic 
happens.

```json
{
	"hipChatFrom": "Upload Magician",
	"hipChatToken": "YoUrToKeNhErE",
	"hipChatRoom": "ROOMID",

	"awsS3Bucket": "YourS3Bucket",

	"urlBase": "https://www.your.com",
	"getRedirect": "/webform.html",

	"badFileExts": ".scr, .exe, .dll",
	"maxFormMemMB": "256",
	"serverHeader": "yes",

	"formNameField": "name",
	"formEmailField": "email",
	"formToField": "contact",
	"formFileField": "file",

	"staticURL": "/static/*",
	"staticPath": "static",

	"formURL": "/static/",
	"thanksURL": "/static/thanks.html"
}
```

If you're running from an EC2 instance in the same region as where your S3 bucket is, 
and the instance that has an IAM role assigned with permissions to "YourS3Bucket", 
you don't need any AWS creds at all. We'll autodetect all the things and just handle it.

**NOTE:** maxFormMemMB doesn't limit the size of the uploads, only how much will be in 
RAM at a time (the rest will get flushed to disk)