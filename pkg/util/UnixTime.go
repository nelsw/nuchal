package util

import (
	"database/sql/driver"
	"errors"
	"time"
	//"strconv"
	"fmt"
	"strconv"
)

type UnixTime struct {
	time.Time
}

func (t UnixTime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(t.UnixNano(), 10)), nil
}
func (t UnixTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return t.Time, nil
}
func (t *UnixTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = UnixTime{Time: value}
		return nil
	}
	return errors.New(fmt.Sprintf("can not convert [%v] to timestamp", v))
}
