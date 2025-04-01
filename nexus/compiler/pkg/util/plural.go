// Copyright 2015 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package util

var consonants = "bcdfghjklmnpqrstvwxyz"
var exceptions = map[string]string{
	// The type name is already in the plural form
	"Endpoints": "Endpoints",
}

// ToPlural returns the plural form of the type's name. If the type's name is found
// in the exceptions map, the map value is returned.
func ToPlural(t string) string {
	singular := t
	var plural string

	if plural, ok := exceptions[singular]; ok {
		return plural
	}

	if len(singular) < 2 {
		return singular
	}

	switch rune(singular[len(singular)-1]) {
	case 's', 'x', 'z':
		plural = esPlural(singular)
	case 'y':
		sl := rune(singular[len(singular)-2])
		if isConsonant(sl) {
			plural = iesPlural(singular)
		} else {
			plural = sPlural(singular)
		}
	case 'h':
		sl := rune(singular[len(singular)-2])
		if sl == 'c' || sl == 's' {
			plural = esPlural(singular)
		} else {
			plural = sPlural(singular)
		}
	case 'e':
		sl := rune(singular[len(singular)-2])
		if sl == 'f' {
			plural = vesPlural(singular[:len(singular)-1])
		} else {
			plural = sPlural(singular)
		}
	case 'f':
		plural = vesPlural(singular)
	default:
		plural = sPlural(singular)
	}
	return plural
}

func iesPlural(singular string) string {
	return singular[:len(singular)-1] + "ies"
}

func vesPlural(singular string) string {
	return singular[:len(singular)-1] + "ves"
}

func esPlural(singular string) string {
	return singular + "es"
}

func sPlural(singular string) string {
	return singular + "s"
}

func isConsonant(char rune) bool {
	for _, c := range consonants {
		if char == c {
			return true
		}
	}
	return false
}
