package customstring

import (
	"errors"
	"strconv"
)

func EnvToInt(str string) (int, error) {
	if str == "" {
		return -1, errors.New("字符串为空")
	}
	integer, err := strconv.Atoi(str)
	if err != nil {
		return -1, err
	}
	return integer, nil
}

func EnvToInt64(str string) (int64, error) {
	if str == "" {
		return -1, errors.New("字符串为空")
	}
	integer64, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return -1, err
	}
	return integer64, nil
}
