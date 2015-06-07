// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gxui

import test "github.com/google/gxui/testing"
import (
	"testing"

	"github.com/google/gxui/interval"
)

func TestTextCursorMergeOne(t *testing.T) {
	s := TextCursor{5, 5}
	l := TextCursorList{}
	interval.Merge(&l, s)
	test.AssertEquals(t, TextCursorList{s}, l)
}

func TestTextCursorMergeInner(t *testing.T) {
	s1 := TextCursor{5, 5}
	s2 := TextCursor{9, -3}
	l := TextCursorList{s1}
	interval.Merge(&l, s2)
	test.AssertEquals(t, TextCursorList{
		TextCursor{10, -5},
	}, l)
}

func TestTextCursorMergeAtStart(t *testing.T) {
	s1 := TextCursor{6, 3}
	s2 := TextCursor{7, -1}
	l := TextCursorList{s1}
	interval.Merge(&l, s2)
	test.AssertEquals(t, TextCursorList{
		TextCursor{9, -3},
	}, l)
}

func TestTextCursorMergeAtEnd(t *testing.T) {
	s1 := TextCursor{6, 3}
	s2 := TextCursor{9, -1}
	l := TextCursorList{s1}
	interval.Merge(&l, s2)
	test.AssertEquals(t, TextCursorList{
		TextCursor{9, -3},
	}, l)
}

func TestTextCursorMergeEncompass(t *testing.T) {
	s1 := TextCursor{9, -3}
	s2 := TextCursor{5, 5}
	l := TextCursorList{s1}
	interval.Merge(&l, s2)
	test.AssertEquals(t, TextCursorList{
		TextCursor{5, 5},
	}, l)
}

func TestTextCursorMergeDuplicate(t *testing.T) {
	s1 := TextCursor{6, -4}
	s2 := TextCursor{2, 4}
	l := TextCursorList{s1}
	interval.Merge(&l, s2)
	test.AssertEquals(t, TextCursorList{
		TextCursor{2, 4},
	}, l)
}

func TestTextCursorMergeDuplicate0Len(t *testing.T) {
	s1 := TextCursor{2, 0}
	s2 := TextCursor{2, 0}
	l := TextCursorList{s1}
	interval.Merge(&l, s2)
	test.AssertEquals(t, TextCursorList{
		TextCursor{2, 0},
	}, l)
}

func TestTextCursorMergeExtendStart(t *testing.T) {
	s1 := TextCursor{9, -3}
	s2 := TextCursor{1, 6}
	l := TextCursorList{s1}
	interval.Merge(&l, s2)
	test.AssertEquals(t, TextCursorList{
		TextCursor{1, 8},
	}, l)
}

func TestTextCursorMergeExtendEnd(t *testing.T) {
	s1 := TextCursor{6, 3}
	s2 := TextCursor{15, -7}
	l := TextCursorList{s1}
	interval.Merge(&l, s2)
	test.AssertEquals(t, TextCursorList{
		TextCursor{15, -9},
	}, l)
}

func TestTextCursorMergeBeforeStart(t *testing.T) {
	s1 := TextCursor{6, 3}
	s2 := TextCursor{6, -4}
	l := TextCursorList{s1}
	interval.Merge(&l, s2)
	test.AssertEquals(t, TextCursorList{
		TextCursor{6, -4},
		TextCursor{6, 3},
	}, l)
}

func TestTextCursorMergeAfterEnd(t *testing.T) {
	s1 := TextCursor{6, -4}
	s2 := TextCursor{6, 3}
	l := TextCursorList{s1}
	interval.Merge(&l, s2)
	test.AssertEquals(t, TextCursorList{
		TextCursor{6, -4},
		TextCursor{6, 3},
	}, l)
}
