package main

import (
	"app/lib/auth/google"
	"app/lib/lock"
	llog "app/lib/log"
	presentation "app/presentation/api"
)

func main() {
	llog.Init()

	if err := google.Init(); err != nil {
		panic(err)
	}

	if err := lock.Init(); err != nil {
		panic(err)
	}

	if err := presentation.Init(); err != nil {
		panic(err)
	}
}
