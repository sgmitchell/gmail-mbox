package gmail

import (
	"fmt"
	"strconv"
	"strings"
)

// MessageIdFromMboxDelim takes the "From " line that delimits messages in mbox files and attempts to parse the gmail
// message id from it.
func MessageIdFromMboxDelim(fromLine string) (string, error) {
	// NB this is about 2x faster than regexp
	idStr, _, _ := strings.Cut(strings.Trim(fromLine, "From "), "@xxx ")
	return IntStrToHexStr(idStr)
}

// IntStrToHexStr takes a string that is an integer and returns a string with its hex value. For some reason, gmail
// uses the hex value for their APIs but uses the int value in their mbox files.
func IntStrToHexStr(intStr string) (string, error) {
	i, err := strconv.ParseInt(intStr, 10, 64)
	if err != nil {
		return "", fmt.Errorf("faild to parse id %q. %w", intStr, err)
	}
	return strconv.FormatInt(i, 16), nil
}
