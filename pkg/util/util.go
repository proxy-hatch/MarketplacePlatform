package util

import (
    "encoding/json"
    "errors"
    "strconv"
    "strings"
    "unicode"
)

func AnyToJsonString(obj any) string {
    return string(AnyToJsonObject(obj))
}
func AnyToJsonObject(obj any) []byte {
    // Convert the Person object to JSON
    jsonData, _ := json.Marshal(obj)
    return jsonData
}

// ConvertPriceStringToInt converts a price string to an int
// For example
// 10 -> 1000
// 10.5 -> 1050
// 12.13 -> 1213
// asdf -> error
// 123.123 -> error
func ConvertPriceStringToInt(s string) (int, error) {
    f, err := strconv.ParseFloat(s, 64)
    if err != nil {
        return 0, err
    }

    parts := strings.Split(s, ".")
    if len(parts) == 2 && len(parts[1]) > 2 {
        return 0, errors.New("more than two decimal places")
    }

    return int(f * 100), nil
}

func SplitArgs(input string) []string {
    var args []string
    var inQuotes bool
    var arg string
    for _, r := range input {
        if r == '\'' {
            inQuotes = !inQuotes
        } else if unicode.IsSpace(r) && !inQuotes {
            args = append(args, arg)
            arg = ""
        } else {
            arg += string(r)
        }
    }
    args = append(args, arg)
    return args
}
