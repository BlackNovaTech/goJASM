package parsers

import (
	"errors"
	"fmt"
	"math/big"
)

func parseInt(strval string, bitlen int) (int64, error) {
	bigval, ok := big.NewInt(-1).SetString(strval, 0)

	if !ok {
		return 0, errors.New(fmt.Sprintf("Invalid value: `%s`", strval))
	}

	if bigval.BitLen() > bitlen {
		return 0, errors.New(fmt.Sprintf("Value out of range: `%s`", strval))
	}

	return bigval.Int64(), nil
}

func ParseInt8(strval string) (int8, error) {
	val, err := parseInt(strval, 8)
	if err != nil {
		return 0, err
	}
	return int8(val), nil
}

func ParseInt16(strval string) (int16, error) {
	val, err := parseInt(strval, 16)
	if err != nil {
		return 0, err
	}
	return int16(val), nil
}

func ParseInt32(strval string) (int32, error) {
	val, err := parseInt(strval, 32)
	if err != nil {
		return 0, err
	}
	return int32(val), nil
}

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
		return 0, errors.New(fmt.Sprintf("Invalid value: `%s`", strval))
	}

	if bigval.BitLen() > bitlen {
		return 0, errors.New(fmt.Sprintf("Value out of range: `%s`", strval))
	}

	if bigval.Sign() < 0 {
		return 0, errors.New(fmt.Sprintf("Value cannot be negative: `%s`", strval))
	}

	return bigval.Uint64(), nil
}

func ParseUint8(strval string) (uint8, error) {
	val, err := parseUint(strval, 8)
	if err != nil {
		return 0, err
	}
	return uint8(val), nil
}

func ParseUint16(strval string) (uint16, error) {
	val, err := parseUint(strval, 16)
	if err != nil {
		return 0, err
	}
	return uint16(val), nil
}

func ParseUint32(strval string) (uint32, error) {
	val, err := parseUint(strval, 32)
	if err != nil {
		return 0, err
	}
	return uint32(val), nil
}

func ParseUint64(strval string) (uint64, error) {
	val, err := parseUint(strval, 64)
	if err != nil {
		return 0, err
	}
	return uint64(val), nil
}
