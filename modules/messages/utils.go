package messages

import (
	"unicode/utf8"

	"github.com/forbole/juno/v4/types"
)

func TrimLastChar(s string) string {
	r, size := utf8.DecodeLastRuneInString(s)
	if r == utf8.RuneError && (size == 0 || size == 1) {
		size = 0
	}
	return s[:len(s)-size]
}

// JoinMessageParsers joins together all the given parsers, calling them in order
func JoinMessageParsers(parsers ...MessageAddressesParser) MessageAddressesParser {
	return func(tx *types.Tx) ([]string, error) {
		for _, parser := range parsers {
			// Try getting the addresses
			addresses, _ := parser(tx)

			// If some addresses are found, return them
			if len(addresses) > 0 {
				return addresses, nil
			}
		}
		return DefaultMessagesParser(tx)
	}
}
