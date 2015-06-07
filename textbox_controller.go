// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gxui

import (
	"sort"
	"strings"
	"unicode"

	"github.com/google/gxui/interval"
	"github.com/google/gxui/math"
)

type TextBoxEdit struct {
	At    int
	Delta int
}

type TextBoxController struct {
	onSelectionChanged          Event
	onTextChanged               Event
	text                        []rune
	lineStarts                  []int
	lineEnds                    []int
	selections                  TextCursorList
	locationHistory             [][]int
	locationHistoryIndex        int
	storeCaretLocationsNextEdit bool
}

func CreateTextBoxController() *TextBoxController {
	t := &TextBoxController{
		onSelectionChanged: CreateEvent(func() {}),
		onTextChanged:      CreateEvent(func([]TextBoxEdit) {}),
	}
	t.selections = TextCursorList{TextCursor{}}
	return t
}

func (t *TextBoxController) textEdited(edits []TextBoxEdit) {
	t.updateSelectionsForEdits(edits)
	t.onTextChanged.Fire(edits)
}

func (t *TextBoxController) updateSelectionsForEdits(edits []TextBoxEdit) {
	min := 0
	max := len(t.text)
	selections := TextCursorList{}
	for _, cursor := range t.selections {
		start, end := cursor.Range()
		for _, e := range edits {
			at := e.At
			delta := e.Delta
			if start > at {
				start += delta
			}
			if end >= at {
				end += delta
			}
		}
		if end < start {
			end = start
		}

		start = math.Clamp(start, min, max)
		end = math.Clamp(end, min, max)

		if cursor.Length > 0 {
			interval.Merge(&selections, CreateTextCursor(end, start))
		} else {
			interval.Merge(&selections, CreateTextCursor(start, end))
		}
	}
	t.selections = selections
}

func (t *TextBoxController) setTextRunesNoEvent(text []rune) {
	t.text = text
	t.lineStarts = t.lineStarts[:0]
	t.lineEnds = t.lineEnds[:0]

	t.lineStarts = append(t.lineStarts, 0)
	for i, r := range text {
		if r == '\n' {
			t.lineEnds = append(t.lineEnds, i)
			t.lineStarts = append(t.lineStarts, i+1)
		}
	}
	t.lineEnds = append(t.lineEnds, len(text))
}

func (t *TextBoxController) maybeStoreCaretLocations() {
	if t.storeCaretLocationsNextEdit {
		t.StoreCaretLocations()
		t.storeCaretLocationsNextEdit = false
	}
}

func (t *TextBoxController) StoreCaretLocations() {
	if t.locationHistoryIndex < len(t.locationHistory) {
		t.locationHistory = t.locationHistory[:t.locationHistoryIndex]
	}
	t.locationHistory = append(t.locationHistory, t.Carets())
	t.locationHistoryIndex = len(t.locationHistory)
}

func (t *TextBoxController) OnSelectionChanged(f func()) EventSubscription {
	return t.onSelectionChanged.Listen(f)
}

func (t *TextBoxController) OnTextChanged(f func([]TextBoxEdit)) EventSubscription {
	return t.onTextChanged.Listen(f)
}

func (t *TextBoxController) SelectionCount() int {
	return len(t.selections)
}

func (t *TextBoxController) Selection(i int) TextCursor {
	return t.selections[i]
}

func (t *TextBoxController) Selections() TextCursorList {
	return append(TextCursorList{}, t.selections...)
}

func (t *TextBoxController) SelectionText(i int) string {
	start, end := t.selections[i].Range()
	runes := t.text[start:end]
	return RuneArrayToString(runes)
}

func (t *TextBoxController) SelectionLineText(i int) string {
	start, _ := t.selections[i].Range()
	line := t.LineIndex(start)
	runes := t.text[t.LineStart(line):t.LineEnd(line)]
	return RuneArrayToString(runes)
}

func (t *TextBoxController) Caret(i int) int {
	return t.selections[i].Index
}

func (t *TextBoxController) Carets() []int {
	l := make([]int, len(t.selections))
	for i, s := range t.selections {
		l[i] = s.Index
	}
	return l
}

func (t *TextBoxController) FirstCaret() int {
	return t.Caret(0)
}

func (t *TextBoxController) LastCaret() int {
	return t.Caret(t.SelectionCount() - 1)
}

func (t *TextBoxController) FirstSelection() TextCursor {
	return t.Selection(0)
}

func (t *TextBoxController) LastSelection() TextCursor {
	return t.Selection(t.SelectionCount() - 1)
}

func (t *TextBoxController) LineCount() int {
	return len(t.lineStarts)
}

func (t *TextBoxController) Line(i int) string {
	return RuneArrayToString(t.LineRunes(i))
}

func (t *TextBoxController) LineRunes(i int) []rune {
	s := t.LineStart(i)
	e := t.LineEnd(i)
	return t.text[s:e]
}

func (t *TextBoxController) LineStart(i int) int {
	if t.LineCount() == 0 {
		return 0
	}
	return t.lineStarts[i]
}

func (t *TextBoxController) LineEnd(i int) int {
	if t.LineCount() == 0 {
		return 0
	}
	return t.lineEnds[i]
}

func (t *TextBoxController) LineIndent(i int) int {
	s, e := t.LineStart(i), t.LineEnd(i)
	l := e - s
	for i := 0; i < l; i++ {
		if !unicode.IsSpace(t.text[i+s]) {
			return i
		}
	}
	return l
}

func (t *TextBoxController) LineIndex(p int) int {
	return sort.Search(len(t.lineStarts), func(i int) bool {
		return p <= t.lineEnds[i]
	})
}

func (t *TextBoxController) Text() string {
	return RuneArrayToString(t.text)
}

func (t *TextBoxController) TextRange(s, e int) string {
	return RuneArrayToString(t.text[s:e])
}

func (t *TextBoxController) TextRunes() []rune {
	return t.text
}

func (t *TextBoxController) SetText(str string) {
	t.SetTextRunes(StringToRuneArray(str))
}

func (t *TextBoxController) SetTextRunes(text []rune) {
	t.setTextRunesNoEvent(text)
	t.textEdited([]TextBoxEdit{})
}

func (t *TextBoxController) SetTextEdits(text []rune, edits []TextBoxEdit) {
	t.setTextRunesNoEvent(text)
	t.textEdited(edits)
}

func (t *TextBoxController) IndexFirst(i int) int {
	return 0
}

func (t *TextBoxController) IndexLast(i int) int {
	return len(t.text)
}

func (t *TextBoxController) IndexLeft(i int) int {
	return math.Max(i-1, 0)
}

func (t *TextBoxController) IndexRight(i int) int {
	return math.Min(i+1, len(t.text))
}

func (t *TextBoxController) IndexWordLeft(i int) int {
	i--
	if i >= 0 {
		wasInWord := t.RuneInWord(t.text[i])
		for i > 0 {
			isInWord := t.RuneInWord(t.text[i-1])
			if isInWord != wasInWord {
				return i
			}
			wasInWord = isInWord
			i--
		}
	}
	return 0
}

func (t *TextBoxController) IndexWordRight(i int) int {
	if i < len(t.text) {
		wasInWord := t.RuneInWord(t.text[i])
		for i < len(t.text)-1 {
			i++
			isInWord := t.RuneInWord(t.text[i])
			if isInWord != wasInWord {
				return i
			}
			wasInWord = isInWord
		}
	}
	return len(t.text)
}

func (t *TextBoxController) IndexUp(i int) int {
	l := t.LineIndex(i)
	x := i - t.LineStart(l)
	if l > 0 {
		return math.Min(t.LineStart(l-1)+x, t.LineEnd(l-1))
	} else {
		return 0
	}
}

func (t *TextBoxController) IndexDown(i int) int {
	l := t.LineIndex(i)
	x := i - t.LineStart(l)
	if l < t.LineCount()-1 {
		return math.Min(t.LineStart(l+1)+x, t.LineEnd(l+1))
	} else {
		return t.LineEnd(l)
	}
}

func (t *TextBoxController) IndexHome(i int) int {
	l := t.LineIndex(i)
	s := t.LineStart(l)
	x := i - s
	indent := t.LineIndent(l)
	if x > indent {
		return s + indent
	} else {
		return s
	}
}

func (t *TextBoxController) IndexEnd(i int) int {
	return t.LineEnd(t.LineIndex(i))
}

type SelectionTransform func(int) int

func (t *TextBoxController) ClearSelections() {
	t.storeCaretLocationsNextEdit = true
	t.SetCaret(t.Caret(0))
}

func (t *TextBoxController) SetCaret(c int) {
	t.storeCaretLocationsNextEdit = true
	t.selections = TextCursorList{}
	t.AddCaret(c)
}

func (t *TextBoxController) AddCaret(c int) {
	t.storeCaretLocationsNextEdit = true
	t.AddSelection(TextCursor{Index: c})
}

func (t *TextBoxController) AddSelection(s TextCursor) {
	t.storeCaretLocationsNextEdit = true
	interval.Merge(&t.selections, s)
	t.onSelectionChanged.Fire()
}

func (t *TextBoxController) SetSelection(s TextCursor) {
	t.storeCaretLocationsNextEdit = true
	t.selections = []TextCursor{s}
	t.onSelectionChanged.Fire()
}

func (t *TextBoxController) SetSelections(s TextCursorList) {
	t.storeCaretLocationsNextEdit = true
	t.selections = s
	if len(s) == 0 {
		t.AddCaret(0)
	} else {
		t.onSelectionChanged.Fire()
	}
}

func (t *TextBoxController) SelectAll() {
	t.storeCaretLocationsNextEdit = true
	t.SetSelection(TextCursor{Index: len(t.text), Length: -len(t.text)})
}

func (t *TextBoxController) RestorePreviousSelections() {
	if t.locationHistoryIndex == len(t.locationHistory) {
		t.StoreCaretLocations()
		t.locationHistoryIndex--
	}
	if t.locationHistoryIndex > 0 {
		t.locationHistoryIndex--
		locations := t.locationHistory[t.locationHistoryIndex]
		t.selections = make(TextCursorList, len(locations))
		for i, l := range locations {
			t.selections[i] = TextCursor{Index: l}
		}
		t.onSelectionChanged.Fire()
	}
}

func (t *TextBoxController) RestoreNextSelections() {
	if t.locationHistoryIndex < len(t.locationHistory)-1 {
		t.locationHistoryIndex++
		locations := t.locationHistory[t.locationHistoryIndex]
		t.selections = make(TextCursorList, len(locations))
		for i, l := range locations {
			t.selections[i] = TextCursor{Index: l}
		}
		t.onSelectionChanged.Fire()
	}
}

func (t *TextBoxController) AddCarets(transform SelectionTransform) {
	t.storeCaretLocationsNextEdit = true
	up := t.selections.Transform(0, transform)
	for _, s := range up {
		interval.Merge(&t.selections, s)
	}
	t.onSelectionChanged.Fire()
}

func (t *TextBoxController) GrowSelections(transform SelectionTransform) {
	t.storeCaretLocationsNextEdit = true
	t.selections = t.selections.TransformRange(0, transform)
	t.onSelectionChanged.Fire()
}

func (t *TextBoxController) MoveSelections(transform SelectionTransform) {
	t.storeCaretLocationsNextEdit = true
	t.selections = t.selections.Transform(0, transform)
	t.onSelectionChanged.Fire()
}

func (t *TextBoxController) AddCaretsUp()       { t.AddCarets(t.IndexUp) }
func (t *TextBoxController) AddCaretsDown()     { t.AddCarets(t.IndexDown) }
func (t *TextBoxController) SelectFirst()       { t.GrowSelections(t.IndexFirst) }
func (t *TextBoxController) SelectLast()        { t.GrowSelections(t.IndexLast) }
func (t *TextBoxController) SelectLeft()        { t.GrowSelections(t.IndexLeft) }
func (t *TextBoxController) SelectRight()       { t.GrowSelections(t.IndexRight) }
func (t *TextBoxController) SelectUp()          { t.GrowSelections(t.IndexUp) }
func (t *TextBoxController) SelectDown()        { t.GrowSelections(t.IndexDown) }
func (t *TextBoxController) SelectHome()        { t.GrowSelections(t.IndexHome) }
func (t *TextBoxController) SelectEnd()         { t.GrowSelections(t.IndexEnd) }
func (t *TextBoxController) SelectLeftByWord()  { t.GrowSelections(t.IndexWordLeft) }
func (t *TextBoxController) SelectRightByWord() { t.GrowSelections(t.IndexWordRight) }
func (t *TextBoxController) MoveFirst()         { t.MoveSelections(t.IndexFirst) }
func (t *TextBoxController) MoveLast()          { t.MoveSelections(t.IndexLast) }
func (t *TextBoxController) MoveLeft()          { t.MoveSelections(t.IndexLeft) }
func (t *TextBoxController) MoveRight()         { t.MoveSelections(t.IndexRight) }
func (t *TextBoxController) MoveUp()            { t.MoveSelections(t.IndexUp) }
func (t *TextBoxController) MoveDown()          { t.MoveSelections(t.IndexDown) }
func (t *TextBoxController) MoveLeftByWord()    { t.MoveSelections(t.IndexWordLeft) }
func (t *TextBoxController) MoveRightByWord()   { t.MoveSelections(t.IndexWordRight) }
func (t *TextBoxController) MoveHome()          { t.MoveSelections(t.IndexHome) }
func (t *TextBoxController) MoveEnd()           { t.MoveSelections(t.IndexEnd) }

func (t *TextBoxController) Delete() {
	t.maybeStoreCaretLocations()
	text := t.text
	edits := []TextBoxEdit{}
	for i := len(t.selections) - 1; i >= 0; i-- {
		s := t.selections[i]
		start, end := s.Range()
		if start == end && end < len(t.text) {
			copy(text[start:], text[start+1:])
			text = text[:len(text)-1]
			edits = append(edits, TextBoxEdit{start, -1})
		} else {
			copy(text[start:], text[end:])
			l := end - start
			text = text[:len(text)-l]
			edits = append(edits, TextBoxEdit{start, -l})
		}
		t.selections[i] = TextCursor{Index: end}
	}
	t.SetTextEdits(text, edits)
}

func (t *TextBoxController) Backspace() {
	t.maybeStoreCaretLocations()
	text := t.text
	edits := []TextBoxEdit{}
	for i := len(t.selections) - 1; i >= 0; i-- {
		s := t.selections[i]
		start, end := s.Range()
		if start == end && start > 0 {
			copy(text[start-1:], text[start:])
			text = text[:len(text)-1]
			edits = append(edits, TextBoxEdit{start - 1, -1})
		} else {
			copy(text[start:], text[end:])
			l := end - start
			text = text[:len(text)-l]
			edits = append(edits, TextBoxEdit{start - 1, -l})
		}
		t.selections[i] = TextCursor{Index: end}
	}
	t.SetTextEdits(text, edits)
}

func (t *TextBoxController) ReplaceAll(str string) {
	t.Replace(func(TextCursor) string { return str })
}

func (t *TextBoxController) ReplaceAllRunes(str []rune) {
	t.ReplaceRunes(func(TextCursor) []rune { return str })
}

func (t *TextBoxController) Replace(f func(sel TextCursor) string) {
	t.ReplaceRunes(func(s TextCursor) []rune { return StringToRuneArray(f(s)) })
}

func (t *TextBoxController) ReplaceRunes(f func(sel TextCursor) []rune) {
	t.maybeStoreCaretLocations()
	text, edit, edits := t.text, TextBoxEdit{}, []TextBoxEdit{}
	for i := len(t.selections) - 1; i >= 0; i-- {
		s := t.selections[i]
		start, end := s.Range()
		text, edit = t.ReplaceAt(text, start, end, f(s))
		edits = append(edits, edit)
	}
	t.setTextRunesNoEvent(text)
	t.textEdited(edits)
}

func (t *TextBoxController) ReplaceAt(text []rune, s, e int, replacement []rune) ([]rune, TextBoxEdit) {
	replacementLen := len(replacement)
	delta := replacementLen - (e - s)
	if delta > 0 {
		text = append(text, make([]rune, delta)...)
	}
	copy(text[e+delta:], text[e:])
	copy(text[s:], replacement)
	if delta < 0 {
		text = text[:len(text)+delta]
	}
	return text, TextBoxEdit{s, delta}
}

func (t *TextBoxController) ReplaceWithNewline() {
	t.ReplaceAll("\n")
	t.Deselect(false)
}

func (t *TextBoxController) ReplaceWithNewlineKeepIndent() {
	t.Replace(func(sel TextCursor) string {
		s, _ := sel.Range()
		indent := t.LineIndent(t.LineIndex(s))
		return "\n" + strings.Repeat(" ", indent)
	})
	t.Deselect(false)
}

func (t *TextBoxController) IndentSelection(tabWidth int) {
	tab := make([]rune, tabWidth)
	for i := range tab {
		tab[i] = ' '
	}
	text, edit, edits := t.text, TextBoxEdit{}, []TextBoxEdit{}
	lastLine := -1
	for i := len(t.selections) - 1; i >= 0; i-- {
		s := t.selections[i]
		start, end := s.Range()
		lis, lie := t.LineIndex(start), t.LineIndex(end)
		if lastLine == lie {
			lie--
		}
		for l := lie; l >= lis; l-- {
			ls := t.LineStart(l)
			text, edit = t.ReplaceAt(text, ls, ls, tab)
			edits = append(edits, edit)
		}
		lastLine = lis
	}
	t.SetTextEdits(text, edits)
}

func (t *TextBoxController) UnindentSelection(tabWidth int) {
	text, edit, edits := t.text, TextBoxEdit{}, []TextBoxEdit{}
	lastLine := -1
	for i := len(t.selections) - 1; i >= 0; i-- {
		s := t.selections[i]
		start, end := s.Range()
		lis, lie := t.LineIndex(start), t.LineIndex(end)
		if lastLine == lie {
			lie--
		}
		for l := lie; l >= lis; l-- {
			c := math.Min(t.LineIndent(l), tabWidth)
			if c > 0 {
				ls := t.LineStart(l)
				text, edit = t.ReplaceAt(text, ls, ls+c, []rune{})
				edits = append(edits, edit)
			}
		}
		lastLine = lis
	}
	t.SetTextEdits(text, edits)
}

func (t *TextBoxController) RuneInWord(r rune) bool {
	switch {
	case unicode.IsLetter(r), unicode.IsNumber(r), r == '_':
		return true
	default:
		return false
	}
}

func (t *TextBoxController) WordAt(runeIdx int) (s, e int) {
	text := t.text
	s, e = runeIdx, runeIdx
	for s > 0 && t.RuneInWord(text[s-1]) {
		s--
	}
	for e < len(t.text) && t.RuneInWord(text[e]) {
		e++
	}
	return s, e
}

func (t *TextBoxController) Deselect(moveCaretToStart bool) (deselected bool) {
	deselected = false
	for i, s := range t.selections {
		if s.Length == 0 {
			continue
		}
		deselected = true
		start, end := s.Range()
		if moveCaretToStart {
			s = TextCursor{Index: start}
		} else {
			s = TextCursor{Index: end}
		}
		t.selections[i] = s
	}
	if deselected {
		t.onSelectionChanged.Fire()
	}
	return
}

func (t *TextBoxController) LineAndRow(index int) (line, row int) {
	line = t.LineIndex(index)
	row = index - t.LineStart(line)
	return
}
