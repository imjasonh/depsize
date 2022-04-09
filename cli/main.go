package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/dustin/go-humanize"
	"github.com/imjasonh/depsize"
)

var (
	human     = flag.Bool("h", false, "if true, humanize bytes")
	recursive = flag.Bool("R", false, "if true, recursively get deps")
)

func main() {
	flag.Parse()
	args := flag.Args()

	dep := args[0]
	var version string
	switch len(args) {
	case 1:
		var err error
		version, err = depsize.Latest(dep)
		if err != nil {
			log.Fatalf("failed to get latest: %v", err)
		}
	case 2:
		version = args[1]
	default:
		log.Fatalf("unexpected args, got %d, want 2", len(args))
	}

	i, err := depsize.Size(dep, version)
	if err != nil {
		log.Fatal(err)
	}

	if *human {
		fmt.Println(dep, version, humanize.Bytes(uint64(i)))
	} else {
		fmt.Println(dep, version, i)
	}

	if *recursive {
		total := i
		deps, err := depsize.Deps(dep, version)
		if err != nil {
			log.Fatalf("getting deps: %v", err)
		}
		for _, d := range deps {
			total += d.Size
			if *human {
				fmt.Println("   ", d.Dep, d.Version, humanize.Bytes(uint64(d.Size)))
			} else {
				fmt.Println("   ", d.Dep, d.Version, d.Size)
			}
		}
		if *human {
			fmt.Println("total", humanize.Bytes(uint64(total)))
		} else {
			fmt.Println("total", total)
		}
	}
}
