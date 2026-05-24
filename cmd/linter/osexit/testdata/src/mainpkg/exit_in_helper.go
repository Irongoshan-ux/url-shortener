package main

import "os"

func helper() {
	os.Exit(1) // want "os.Exit outside main"
}
