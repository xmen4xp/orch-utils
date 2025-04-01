// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	re "regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Save repeated regexp compilation
func Test_compileRegexp(t *testing.T) {
	vrex := new(re.Regexp)

	rex, err := compileRegexp(".*TestRegexp.*")
	assert.NoError(t, err)
	assert.NotNil(t, rex)
	assert.IsType(t, vrex, rex)

	rex, err = compileRegexp(".*TestRegexp.*")
	assert.NoError(t, err)
	assert.NotNil(t, rex)

	irex, ierr := compileRegexp(".[*InvalidTestRegexp.*")
	assert.Error(t, ierr)
	assert.Nil(t, irex)
	assert.IsType(t, vrex, irex)
}

// Save repeated regexp compilation, with panic on error
func testPanic() {
	mustCompileRegexp(".[*InvalidTestRegexp.*")
}

func Test_mustCompileRegexp(t *testing.T) {
	vrex := new(re.Regexp)

	rex := mustCompileRegexp(".*TestRegexp2.*")
	assert.NotNil(t, rex)
	assert.IsType(t, vrex, rex)

	rex = mustCompileRegexp(".*TestRegexp2.*")
	assert.NotNil(t, rex)

	assert.Panics(t, testPanic)
}

func TestRace_compileRegexp(t *testing.T) {
	vrex := new(re.Regexp)

	patterns := []string{
		".*TestRegexp1.*",
		".*TestRegexp2.*",
		".*TestRegexp3.*",
	}

	comp := func(pattern string) {
		rex, err := compileRegexp(pattern)
		assert.NoError(t, err)
		assert.NotNil(t, rex)
		assert.IsType(t, vrex, rex)
	}

	for i := 0; i < 20; i++ {
		t.Run(patterns[i%3], func(t *testing.T) {
			t.Parallel()
			comp(patterns[i%3])
		})
	}
}

func TestRace_mustCompileRegexp(t *testing.T) {
	vrex := new(re.Regexp)

	patterns := []string{
		".*TestRegexp1.*",
		".*TestRegexp2.*",
		".*TestRegexp3.*",
	}

	comp := func(pattern string) {
		rex := mustCompileRegexp(pattern)
		assert.NotNil(t, rex)
		assert.IsType(t, vrex, rex)
	}

	for i := 0; i < 20; i++ {
		t.Run(patterns[i%3], func(t *testing.T) {
			t.Parallel()
			comp(patterns[i%3])
		})
	}
}
