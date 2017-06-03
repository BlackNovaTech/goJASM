package parsers

import (
	"fmt"
	"math/big"
	"strings"
)

func parseInt(strval string, bitlen int) (int64, error) {
	bigval, ok := big.NewInt(-1).SetString(strval, 0)

	if !ok {
		return 0, fmt.Errorf("Invalid value: `%s`", strval)
	}

	if bigval.BitLen() > bitlen {
		return 0, fmt.Errorf("Value out of range: `%s`", strval)
	}

	return bigval.Int64(), nil
}

// ParseInt8 parses an int8 in binary, octal, decimal or hex from the given string
func ParseInt8(strval string) (int8, error) {
	val, err := parseInt(strval, 8)
	if err != nil {
		return 0, err
	}
	return int8(val), nil
}

// ParseInt16 parses an int16 in binary, octal, decimal or hex from the given string
func ParseInt16(strval string) (int16, error) {
	val, err := parseInt(strval, 16)
	if err != nil {
		return 0, err
	}
	return int16(val), nil
}

// ParseInt32 parses an int32 in binary, octal, decimal or hex from the given string
func ParseInt32(strval string) (int32, error) {
	val, err := parseInt(strval, 32)
	if err != nil {
		return 0, err
	}
	return int32(val), nil
}

// ParseInt64 parses an int64 in binary, octal, decimal or hex from the given string
func ParseInt64(strval string) (int64, error) {
	val, err := parseInt(strval, 64)
	if err != nil {
		return 0, err
	}
	return int64(val), nil
}

func parseUint(strval string, bitlen int) (uint64, error) {
	bigval, ok := big.NewInt(-1).SetString(strval, 0)

	if !ok {
		return 0, fmt.Errorf("Invalid value: `%s`", strval)
	}

	if bigval.BitLen() > bitlen {
		return 0, fmt.Errorf("Value out of range: `%s`", strval)
	}

	if bigval.Sign() < 0 {
		return 0, fmt.Errorf("Value cannot be negative: `%s`", strval)
	}

	return bigval.Uint64(), nil
}

// ParseUint8 parses an uint8 in binary, octal, decimal or hex from the given string
func ParseUint8(strval string) (uint8, error) {
	val, err := parseUint(strval, 8)
	if err != nil {
		return 0, err
	}
	return uint8(val), nil
}

// ParseUint16 parses an uint16 in binary, octal, decimal or hex from the given string
func ParseUint16(strval string) (uint16, error) {
	val, err := parseUint(strval, 16)
	if err != nil {
		return 0, err
	}
	return uint16(val), nil
}

// ParseUint32 parses an uint32 in binary, octal, decimal or hex from the given string
func ParseUint32(strval string) (uint32, error) {
	val, err := parseUint(strval, 32)
	if err != nil {
		return 0, err
	}
	return uint32(val), nil
}

// ParseUint64 parses an uint64 in binary, octal, decimal or hex from the given string
func ParseUint64(strval string) (uint64, error) {
	val, err := parseUint(strval, 64)
	if err != nil {
		return 0, err
	}
	return uint64(val), nil
}

// ParseChar parses an uint8 from a character literal
func ParseChar(strval string) (int8, error) {
	if strings.HasPrefix(strval, "'") {
		if !strings.HasSuffix(strval, "'") || len(strval) != 3 {
			return 0, fmt.Errorf("Invalid character literal `%s`", strval)
		}
		return int8(strval[1]), nil
	}
	return 0, fmt.Errorf("Invalid character literal `%s`", strval)
}