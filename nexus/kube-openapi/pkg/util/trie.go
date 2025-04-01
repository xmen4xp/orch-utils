// Copyright 2016 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package util

// A simple trie implementation with Add and HasPrefix methods only.
type Trie struct {
	children map[byte]*Trie
	wordTail bool
	word     string
}

// NewTrie creates a Trie and add all strings in the provided list to it.
func NewTrie(list []string) Trie {
	ret := Trie{
		children: make(map[byte]*Trie),
		wordTail: false,
	}
	for _, v := range list {
		ret.Add(v)
	}
	return ret
}

// Add adds a word to this trie
func (t *Trie) Add(v string) {
	root := t
	for _, b := range []byte(v) {
		child, exists := root.children[b]
		if !exists {
			child = &Trie{
				children: make(map[byte]*Trie),
				wordTail: false,
			}
			root.children[b] = child
		}
		root = child
	}
	root.wordTail = true
	root.word = v
}

// HasPrefix returns true of v has any of the prefixes stored in this trie.
func (t *Trie) HasPrefix(v string) bool {
	_, has := t.GetPrefix(v)
	return has
}

// GetPrefix is like HasPrefix but return the prefix in case of match or empty string otherwise.
func (t *Trie) GetPrefix(v string) (string, bool) {
	root := t
	if root.wordTail {
		return root.word, true
	}
	for _, b := range []byte(v) {
		child, exists := root.children[b]
		if !exists {
			return "", false
		}
		if child.wordTail {
			return child.word, true
		}
		root = child
	}
	return "", false
}
