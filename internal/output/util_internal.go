package output

import "strconv"

func intToString(i int) string   { return strconv.Itoa(i) }
func boolToString(b bool) string { return strconv.FormatBool(b) }
