package errror

import (
	"fmt"
	"log"
	"runtime"
	"strings"
)

// we WRAP the error exactly where it appears and then we pass it up
// if after its transported up it need to be sent to admin, we ErrInJson it

// returns error for logging. >
func FormatL(err0 error) error {
	return fmt.Errorf(">%s",
		formatError(err0))
}

// returns user error. &
func FormatU(err0 error) error {
	return fmt.Errorf("&%s",
		formatError(err0))
}

// returns user error. &
func FormatsU(err0 string) error {
	return fmt.Errorf("&%s",
		formatError(fmt.Errorf(err0)))
}

// returns error for admin. !
func FormatA(err0 error) error {
	return fmt.Errorf("!%s",
		formatError(err0))
}

func formatError(err0 error) string {
	_, file, cl, ok := runtime.Caller(2)
	if !ok {
		log.Println("COULDNT WRAP ERROR")
	}

	fileSep := strings.Split(file, "/")
	if len(fileSep) > 2 {
		file = fileSep[len(fileSep)-2]
	}

	return fmt.Sprintf("%s LINE %d: %s", file, cl, err0)
}
