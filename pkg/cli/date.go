package cli

import (
	"fmt"
	"time"
)

func RunDate() error {
	fmt.Println(time.Now().Format(time.RFC3339))
	return nil
}
