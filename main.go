package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

var (
	//directory where Jekyll will look to transform files
	source = flag.String("source", "", "")

	// directory where Jekyll will write files to
	destination = flag.String("destination", "_site", "")

	// fires up a server that will host your _site directory if True
	server = flag.Bool("server", false, "")

	// the port that the Jekyll server will run on
	port = flag.String("server_port", ":4000", "")

	// runs Jekyll with verbose output if True
	verbose = flag.Bool("verbose", false, "")

	// displays the help / usage if True
	help = flag.Bool("help", false, "")
)

func main() {

	// Parse the input parameters
	flag.BoolVar(help, "h", false, "")
	flag.BoolVar(verbose, "v", false, "")
	flag.Usage = usage
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// User may specify the source as a non-flag variable
	if flag.NArg() > 0 {
		source = &flag.Args()[0]
	}

	// Convert the directory to an absolute path
	src, _ := filepath.Abs(*source)
	dest, _ := filepath.Abs(*destination)

	// Change the working directory to the website's source directory
	os.Chdir(src)

	// Initialize the Jekyll website
	site, err := NewSite(src, dest)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Generate the static website
	if err := site.Generate(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// If the server option is enabled, launch a webserver
	if *server {

		// Change the working directory to the _site directory
		os.Chdir(dest)

		// Create the handler to serve from the filesystem
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			path := filepath.Clean(r.URL.Path)
			path = filepath.Join(dest, path)
			http.ServeFile(w, r, path)
		})

		// Serve the website from the _site directory
		fmt.Printf("Starting server on port %s\n", *port)
		if err := http.ListenAndServe(*port, nil); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	os.Exit(0)
}

var usage = func() {
	fmt.Println(`Usage: jkl [OPTION]... [SOURCE]

      --source         changes the dir where Jekyll will look to transform files
      --destination    changes the dir where Jekyll will write files to
      --server         starts a server that will host your _site directory
      --server-port    changes the port that the Jekyll server will run on
  -v, --verbose        runs Jekyll with verbose output
  -h, --help           display this help and exit

Examples:
  jkl                       generates site from current working dir
  jkl --server              generates site and serves at localhost:4000
  jkl /home/ubutnu/mysite   generates site from source dir /home/ubuntu/mysite

Report bugs to <https://github.com/bradrydzewski/jkl/issues>
Jekyll home page: <https://github.com/bradrydzewski/jkl>`)
}
