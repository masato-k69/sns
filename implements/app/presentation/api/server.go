package presentation

import (
	"app/lib/echo/session"
	"app/presentation/subscriber"
	"os"
	"sync"
)

func Init() error {
	if err := session.Init(); err != nil {
		return err
	}

	router, err := NewRouter()

	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		subscriber.Start()
	}()
	go func() {
		defer wg.Done()
		router.Logger.Fatal(router.Start(":" + os.Getenv("SERVER_HOST")))
	}()
	wg.Wait()

	return nil
}
