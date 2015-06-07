// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gxui

import "github.com/google/gxui/interval"

// TextCursor represents a single point in a string. When Length is less than
// or greater than zero, the TextCursor represents an interval of runes, where
// Index and Index + Length can be considered as the indices to the gaps between
// runes.
//
// For example, given the string "Hello world":
//
//      ┌   ┬   ┬   ┬   ┬   ┬   ┬   ┬   ┬   ┬   ┬   ┐
//        H   e   l   l   o       w   o   r   l   d
//      └   ┴   ┴   ┴   ┴   ┴   ┴   ┴   ┴   ┴   ┴   ┘
//      ₀   ₁   ₂   ₃   ₄   ₅   ₆   ₇   ₈   ₉   ₁₀  ₁₁
//
//
// TextCursor{ Index: 5, Length: -5 } represents:
//
//      ┌   ┬   ┬   ┬   ┬   ╥   ┬   ┬   ┬   ┬   ┬   ┐
//      │ H   e   l   l   o ║     w   o   r   l   d
//      └   ┴   ┴   ┴   ┴   ╨   ┴   ┴   ┴   ┴   ┴   ┘
//      ├───────────────────╢
//      ₀                   ₅
//
//
// TextCursor{ Index: 6, Length: 5 } represents:
//
//      ┌   ┬   ┬   ┬   ┬   ┬   ╥   ┬   ┬   ┬   ┬   ┐
//        H   e   l   l   o     ║ w   o   r   l   d
//      └   ┴   ┴   ┴   ┴   ┴   ╨   ┴   ┴   ┴   ┴   ┘
//                              ╟───────────────────┤
//                              ₆                   ₁₁
//
//
// TextCursor{ Index: 5, Length: 0 } represents:
//
//      ┌   ┬   ┬   ┬   ┬   ╥   ┬   ┬   ┬   ┬   ┬   ┐
//        H   e   l   l   o ║     w   o   r   l   d
//      └   ┴   ┴   ┴   ┴   ╨   ┴   ┴   ┴   ┴   ┴   ┘
//                          ║
//                          ₅
type TextCursor struct {
	Index  int // Caret index.
	Length int // If non-zero, cursor is a selection.
}

// CreateTextCursor is a helper function that returns {Index: to, Length: from - to}.
func CreateTextCursor(from, to int) TextCursor {
	return TextCursor{to, from - to}
}

// Range returns the cursor indices ordered least to greatest.
func (c TextCursor) Range() (int, int) {
	n := c.Index + c.Length
	if c.Index < n {
		return c.Index, n
	}
	return n, c.Index
}

// Invert returns the cursor in opposite order.
func (c TextCursor) Invert() TextCursor {
	return TextCursor{
		Index:  c.Index + c.Length,
		Length: -c.Length,
	}
}

// CaretAtStart returns true if the cursor index is at the start of a selection.
// If cursor length is zero, returns false.
func (c TextCursor) CaretAtStart() bool {
	return c.Length > 0
}

// Span implements the interval.Node interface.
func (c TextCursor) Span() (uint64, uint64) {
	start, end := c.Range()
	return uint64(start), uint64(end)
}

// TextCursorList attaches the methods of interval.List, interval.RList,
// and interval.ExtendedList to support bulk transforms.
type TextCursorList []TextCursor

// Transform alters cursor index to change location.
func (p TextCursorList) Transform(from int, transform func(i int) int) TextCursorList {
	res := TextCursorList{}
	for _, c := range p {
		if c.Index >= from {
			c.Index = transform(c.Index)
		}
		interval.Merge(&res, c)
	}
	return res
}

// TransformRange alters cursor to increase or decrease range.
func (p TextCursorList) TransformRange(from int, transform func(i int) int) TextCursorList {
	res := TextCursorList{}
	for _, c := range p {
		if c.Index >= from {
			idx := transform(c.Index)
			c.Length += c.Index - idx
			c.Index = idx
		}
		interval.Merge(&res, c)
	}
	return res
}

// Len implements interval.RList interface.
func (p TextCursorList) Len() int { return len(p) }

// Cap implements interval.List interface.
func (p TextCursorList) Cap() int { return cap(p) }

// SetLen implements interval.List interface.
func (p *TextCursorList) SetLen(len int) {
	*p = (*p)[:len]
}

// GrowTo implements interval.List interface.
func (p *TextCursorList) GrowTo(length, capacity int) {
	old := *p
	*p = make(TextCursorList, length, capacity)
	copy(*p, old)
}

// Copy implements interval.List interface.
func (p TextCursorList) Copy(to, from, count int) {
	copy(p[to:to+count], p[from:from+count])
}

// GetInterval implements interval.RList interface.
func (p TextCursorList) GetInterval(index int) (start, end uint64) {
	return p[index].Span()
}

// SetInterval implements interval.RList interface. Creates new selection
// oriented as a previous cursor at index, or an empty cursor otherwise.
func (p TextCursorList) SetInterval(index int, start, end uint64) {
	p[index] = CreateTextCursor(int(start), int(end))
}

// MergeData implements interval.ExtendedList interface. Orients cursor at
// index to i.
func (p TextCursorList) MergeData(index int, i interval.Node) {
	if i.(TextCursor).CaretAtStart() != p[index].CaretAtStart() {
		p[index] = p[index].Invert()
	}
}
