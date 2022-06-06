package closer

import (
	"fmt"
	"time"
)

// ExampleClose is an example of how to use closer in your main func.
func ExampleClose() {
	Bind(cleanupFunc)

	go func() {
		// do some pseudo background work
		fmt.Println("10 seconds to go...")
		time.Sleep(10 * time.Second)

		Close()
	}()

	Hold()
}

func cleanupFunc() {
	fmt.Print("Hang on! I'm closing some DBs, wiping some trails..")
	time.Sleep(3 * time.Second)
	fmt.Println("  Done.")
}
