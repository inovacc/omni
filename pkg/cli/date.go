package cli

import (
	"fmt"
	"github.com/inovacc/goshell/pkg/timeutil"
)

func RunDate() error {
	fmt.Println(timeutil.Date(""))
	return nil
}
