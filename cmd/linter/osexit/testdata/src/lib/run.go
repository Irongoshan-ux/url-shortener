package lib

import "log"

func Run() {
	log.Fatal("stop") // want "log.Fatal outside main"
}
