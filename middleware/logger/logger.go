package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/sattvikc/go-simpleapi"
)

const reset = "\033[0m"
const bold = "\033[1m"
const underline = "\033[4m"
const strike = "\033[9m"
const italic = "\033[3m"

const cRed = "\033[31m"
const cGreen = "\033[32m"
const cYellow = "\033[33m"
const cBlue = "\033[34m"
const cPurple = "\033[35m"
const cCyan = "\033[36m"
const cWhite = "\033[37m"

func New() interface{} {
	return func(ctx *simpleapi.Context) error {

		startTime := time.Now()
		err := ctx.Next()
		endTime := time.Now()

		timeTaken := endTime.Sub(startTime)

		fmt.Fprintf(
			os.Stderr,
			"%s | %d | %s | %s | %s | %s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			200, // response status
			fomatDuration(timeTaken),
			ctx.Request.RemoteAddr,
			cGreen+ctx.Request.Method+reset,
			cYellow+ctx.Request.URL.Path+reset,
		)
		return err
	}
}

func fomatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	} else if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}
