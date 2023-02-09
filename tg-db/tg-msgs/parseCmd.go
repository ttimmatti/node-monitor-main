package msgs

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ttimmatti/nodes-bot/tg-db/errror"
)

func validServerIp(server_ip string) bool {
	// x.x.x.x, 0 <= x <= 255
	arrX := strings.Split(server_ip, ".")

	if len(arrX) > 4 {
		return false
	}

	for _, xS := range arrX {
		x, err := strconv.Atoi(xS)
		if err != nil {
			return false
		}

		if x < 0 || x > 255 {
			return false
		}
	}

	return true
}

// - SERVERS: _, _, _, err
//
// - ADD/DELETE: cmd, server_ip, _, err
//
// - UPDATE: cmd, server_ip, server_ip, err
func parseCmd(text string) (string, string, string, error) {
	//parses the command

	//if it's not a cmd
	if !strings.HasPrefix(text, "/") {
		return "", "", "", fmt.Errorf("not a cmd")
	}

	textS := strings.Split(text, " ")
	textN := len(textS)
	if len(textS) < 1 {
		//TODO: let me know about this error
		return "", "", "", fmt.Errorf("!FOR SOME REASON TEXT WAS EMPTY!!! returning")
	}

	// for /start and /delete we need only cmd and value
	// for /update we need cmd and two values
	// for /read we need no values
	switch textN {
	case 1:
		// only cmd -- read
		return textS[0], "", "", nil
	case 2:
		// cmd and val -- add/delete
		return textS[0], textS[1], "", nil
	case 3:
		// cmd and 2 vals -- update
		return textS[0], textS[1], textS[2], nil
	default:
		return "", "", "", errror.NewErrorf(
			errror.ErrorCodeWrongCmd,
			"wrongCmd")
		//TODO: RETURN ERROR TO USER
	}
}
