// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package strfmt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDate(t *testing.T) {
	pp := Date{}
	err := pp.UnmarshalText([]byte{})
	assert.NoError(t, err)
	err = pp.UnmarshalText([]byte("yada"))
	assert.Error(t, err)

	orig := "2014-12-15"
	bj := []byte("\"" + orig + "\"")
	err = pp.UnmarshalText([]byte(orig))
	assert.NoError(t, err)

	txt, err := pp.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, orig, string(txt))

	err = pp.UnmarshalJSON(bj)
	assert.NoError(t, err)
	assert.EqualValues(t, orig, pp.String())

	err = pp.UnmarshalJSON([]byte(`"1972/01/01"`))
	assert.Error(t, err)

	b, err := pp.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, bj, b)

	var dateZero Date
	err = dateZero.UnmarshalJSON([]byte(jsonNull))
	assert.NoError(t, err)
	assert.Equal(t, Date{}, dateZero)
}

func TestDate_IsDate(t *testing.T) {
	tests := []struct {
		value string
		valid bool
	}{
		{"2017-12-22", true},
		{"2017-1-1", false},
		{"17-13-22", false},
		{"2017-02-29", false}, // not a valid date : 2017 is not a leap year
		{"1900-02-29", false}, // not a valid date : 1900 is not a leap year
		{"2100-02-29", false}, // not a valid date : 2100 is not a leap year
		{"2000-02-29", true},  // a valid date : 2000 is a leap year
		{"2400-02-29", true},  // a valid date : 2000 is a leap year
		{"2017-13-22", false},
		{"2017-12-32", false},
		{"20171-12-32", false},
		{"YYYY-MM-DD", false},
		{"20-17-2017", false},
		{"2017-12-22T01:02:03Z", false},
	}
	for _, test := range tests {
		assert.Equal(t, test.valid, IsDate(test.value), "value [%s] should be valid: [%t]", test.value, test.valid)
	}
}

func TestDeepCopyDate(t *testing.T) {
	ref := time.Now().Truncate(24 * time.Hour).UTC()
	date := Date(ref)
	in := &date

	out := new(Date)
	in.DeepCopyInto(out)
	assert.Equal(t, in, out)

	out2 := in.DeepCopy()
	assert.Equal(t, in, out2)

	var inNil *Date
	out3 := inNil.DeepCopy()
	assert.Nil(t, out3)
}
