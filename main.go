package main

/*
This is largely built to run via AWS Elastic Beanstalk.
The following environment variables are helpful:

DEBUG=true
PORT=<port to listen on>
AWS_ACCESS_KEY=<AWS access key id>
AWS_SECRET_KEY=<AWS secret key>
AWS_REGION=<AWS region used>

CONFIGFOLDER=<relative path to config folder>

*/

import (
	"flag"
	"fmt"
	"github.com/zenazn/goji"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
)

var (
	SHORTVERSION string      = "1.0.0"
	NAME         string      = "H23"
	FULLVERSION  string      = fmt.Sprintf("%s/%s %s %d:%d", NAME, SHORTVERSION, runtime.Version(), runtime.GOMAXPROCS(0), runtime.NumCPU())
	debugOut     *log.Logger = log.New(ioutil.Discard, "[DEBUG]", log.Lshortfile)
	debug        bool
	configFolder string
	GlobalConfig Config = make(Config)
)

func init() {
	// Misc options
	flag.BoolVar(&debug, "debug", false, "Enable Debug output")
	flag.StringVar(&configFolder, "configfolder", "configs/", "Where the config files live")
	GlobalConfig.PSet("tempFolderRoot", flag.String("temproot", os.TempDir(), "Where to create temp folders and files"))
	GlobalConfig.PSet("getRedirect", flag.String("getredirect", "", "On GET (vs. POST) where to redirect the misguided to"))

	// HipChat options
	GlobalConfig.PSet("hipChatFrom", flag.String("from", "", "HipChat username to send from"))
	GlobalConfig.PSet("hipChatToken", flag.String("token", "", "HipChat token to use"))
	GlobalConfig.PSet("hipChatRoom", flag.String("room", "", "HipChat room to send to"))

	// AWS options
	GlobalConfig.PSet("awsRegion", flag.String("awsregion", "", "AWS region"))
	GlobalConfig.PSet("awsAccessKey", flag.String("awsaccesskey", "", "AWS access key"))
	GlobalConfig.PSet("awsSecretKey", flag.String("awssecretkey", "", "AWS secret key"))
	GlobalConfig.PSet("awsS3Bucket", flag.String("bucket", "", "S3 Bucket to put files in"))

	// Goji-mandated
	flag.Set("bind", ":"+os.Getenv("PORT"))

	flag.Parse()
}

func main() {

	// Handle debugging
	if debug || os.Getenv("DEBUG") == "true" {
		debug = true // Just in case
		debugOut = log.New(os.Stdout, "[DEBUG]", log.Lshortfile)
	}

	debugOut.Printf("Pre-Config:\n%+v\n", GlobalConfig.Map())

	// Load Configs
	if configFolder != "" {
		loadConfigs(configFolder)
	} else if cf := os.Getenv("CONFIGFOLDER"); cf != "" {
		loadConfigs(cf)
	}

	debugOut.Printf("Post-Config\n%+v\n", GlobalConfig.Map())

	// Setup AWS stuff
	initAWS()

	// Goji!!!
	if GlobalConfig.IsNotNull("serverHeader") {
		headerString := GlobalConfig.Get("serverHeader")
		if headerString != "yes" {
			FULLVERSION = headerString
		}
		goji.Use(ServerHeader)
	}

	goji.Get("/", http.RedirectHandler(GlobalConfig.Get("formURL"), 301))
	goji.Get("/health", healthHandler)
	goji.Post("/upload", uploadHandler)
	goji.Get("/upload", http.RedirectHandler(GlobalConfig.Get("getRedirect"), 301))

	// Allow handling of static content for webform, thank you page, etc.
	if GlobalConfig.IsNotNull("staticPath") && GlobalConfig.IsNotNull("staticURL") {
		debugOut.Printf("Static handling of '%s' mapped to '%s'\n", GlobalConfig.Get("staticURL"), GlobalConfig.Get("staticPath"))
		goji.Handle(GlobalConfig.Get("staticURL"),
			http.StripPrefix(strings.TrimRight(GlobalConfig.Get("staticURL"), "*"),
				http.FileServer(http.Dir(GlobalConfig.Get("staticPath")))))
	}

	goji.Handle("/*", defaultHandler)

	goji.Serve()
}
