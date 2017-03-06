// Copyright (c) 2017 CHEN Xianren. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package query

import (
	"fmt"
)

const (
	OptionZeroValue     = 0
	OptionAutoIncrement = 1 << iota
	OptionAutoNow
	OptionAutoNowAdd
	OptionVersion
)

type Starter interface {
	Dialect() string
	Quote(string) string
	Quoted(string) string
	Parameter(bool, int) string
	Returning(byte, string) string
	Mapping(string, int, int) (string, string)
}

func NewStarter(dialect string) Starter {
	switch dialect {
	case "mysql":
		return MySQLStarter
	case "postgres":
		return PostgreSQLStarter
	case "sqlite":
		return SQLiteStarter
	default:
		return StandardStarter
	}
}

type Standard struct{}

var (
	StandardStarter         = Standard{}
	_               Starter = StandardStarter
)

func (Standard) Dialect() string {
	return "standard"
}

func (Standard) Parameter(n bool, i int) string {
	if n {
		return ""
	} else {
		return "?"
	}
}

func (Standard) Quote(s string) string {
	if len(s) == 0 || len(s) > maxLen {
		return ""
	}
	for _, r := range s {
		if notAllow(r) {
			return ""
		}
	}
	return `"` + s + `"`
}

func (standard Standard) Quoted(s string) string {
	if s[0] == '\'' {
		return s
	} else {
		return standard.Quote(s[1 : len(s)-1])
	}
}

func (Standard) Returning(byte, string) string {
	return ""
}

func (Standard) Mapping(goType string, maxSize, option int) (_, optionValue string) {
	switch option {
	case OptionAutoIncrement:
		optionValue = "GENERATED BY DEFAULT AS IDENTITY"
	case OptionAutoNow, OptionAutoNowAdd:
		if goType == "time" {
			optionValue = "DEFAULT CURRENT_TIMESTAMP"
		} else {
			optionValue = "DEFAULT 0"
		}
	case OptionVersion:
		optionValue = "DEFAULT 1"
	}
	switch goType {
	case "bool":
		return "BOOLEAN", "FALSE"
	case "int8", "int16", "uint8", "uint16":
		if option == OptionZeroValue {
			optionValue = "0"
		}
		return "SMALLINT", optionValue
	case "int32", "int64", "int", "uint32", "uint64", "uint":
		if option == OptionZeroValue {
			optionValue = "0"
		}
		return "INTEGER", optionValue
	case "float32":
		return "REAL", "0"
	case "float64":
		return "DOUBLE PRECISION", "0"
	case "time":
		if option == OptionZeroValue {
			optionValue = "'1970-01-01T00:00:00Z'"
		}
		return "TIMESTAMP WITH TIME ZONE", optionValue
	case "bytes", "gob":
		if maxSize > 0 && maxSize <= 255 {
			return fmt.Sprintf("BINARY LARGE OBJECT(%d)", maxSize), ""
		} else {
			return "BINARY LARGE OBJECT", ""
		}
	case "string": // interface, json, xml
		optionValue = "''"
		fallthrough
	default:
		if maxSize == 0 {
			return "CHARACTER VARYING(255)", optionValue
		} else if maxSize > 0 && maxSize <= 255 {
			return fmt.Sprintf("CHARACTER VARYING(%d)", maxSize), optionValue
		} else {
			return "CHARACTER LARGE OBJECT", optionValue
		}
	}
}
