package rangeset

import (
	"slices"
	"testing"
)

var defaultStart = []RangeEntry[string]{{"f.x.", "h.x."}, {"m.x.", "t.x."}, {"z.x.", "x."}}

// TODO move these tests to number ranges instead to avoid the import lol
func TestAddRanges(t *testing.T) {
	for _, datum := range []struct {
		start    []RangeEntry[string]
		in       RangeEntry[string]
		expected []RangeEntry[string]
	}{
		// entirely before
		{in: RangeEntry[string]{"a.x.", "b.x."}, expected: []RangeEntry[string]{{"a.x.", "b.x."}, {"f.x.", "h.x."}, {"m.x.", "t.x."}, {"z.x.", "x."}}},
		// starts before, ends within first range
		{in: RangeEntry[string]{"a.x.", "f.x."}, expected: []RangeEntry[string]{{"a.x.", "h.x."}, {"m.x.", "t.x."}, {"z.x.", "x."}}},
		{in: RangeEntry[string]{"a.x.", "g.x."}, expected: []RangeEntry[string]{{"a.x.", "h.x."}, {"m.x.", "t.x."}, {"z.x.", "x."}}},
		{in: RangeEntry[string]{"a.x.", "h.x."}, expected: []RangeEntry[string]{{"a.x.", "h.x."}, {"m.x.", "t.x."}, {"z.x.", "x."}}},
		// starts before, ends between first and second
		{in: RangeEntry[string]{"a.x.", "i.x."}, expected: []RangeEntry[string]{{"a.x.", "i.x."}, {"m.x.", "t.x."}, {"z.x.", "x."}}},
		// starts before, ends within second range
		{in: RangeEntry[string]{"a.x.", "m.x."}, expected: []RangeEntry[string]{{"a.x.", "t.x."}, {"z.x.", "x."}}},
		{in: RangeEntry[string]{"a.x.", "n.x."}, expected: []RangeEntry[string]{{"a.x.", "t.x."}, {"z.x.", "x."}}},
		{in: RangeEntry[string]{"a.x.", "t.x."}, expected: []RangeEntry[string]{{"a.x.", "t.x."}, {"z.x.", "x."}}},
		// starts before first, ends after last range
		{in: RangeEntry[string]{"a.x.", "v.x."}, expected: []RangeEntry[string]{{"a.x.", "v.x."}, {"z.x.", "x."}}},

		// starts and ends in first
		{in: RangeEntry[string]{"f.x.", "g.x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"m.x.", "t.x."}, {"z.x.", "x."}}},
		{in: RangeEntry[string]{"f.x.", "h.x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"m.x.", "t.x."}, {"z.x.", "x."}}},
		// starts in first, ends between first and second
		{in: RangeEntry[string]{"f.x.", "i.x."}, expected: []RangeEntry[string]{{"f.x.", "i.x."}, {"m.x.", "t.x."}, {"z.x.", "x."}}},

		// between first and second
		{in: RangeEntry[string]{"i.x.", "j.x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"i.x.", "j.x."}, {"m.x.", "t.x."}, {"z.x.", "x."}}},

		// starts between, ends in second
		{in: RangeEntry[string]{"i.x.", "m.x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"i.x.", "t.x."}, {"z.x.", "x."}}},
		{in: RangeEntry[string]{"i.x.", "n.x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"i.x.", "t.x."}, {"z.x.", "x."}}},
		{in: RangeEntry[string]{"i.x.", "t.x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"i.x.", "t.x."}, {"z.x.", "x."}}},

		// starts in first, ends in second
		{in: RangeEntry[string]{"f.x.", "t.x."}, expected: []RangeEntry[string]{{"f.x.", "t.x."}, {"z.x.", "x."}}},
		{in: RangeEntry[string]{"g.x.", "t.x."}, expected: []RangeEntry[string]{{"f.x.", "t.x."}, {"z.x.", "x."}}},
		{in: RangeEntry[string]{"h.x.", "t.x."}, expected: []RangeEntry[string]{{"f.x.", "t.x."}, {"z.x.", "x."}}},
		{in: RangeEntry[string]{"f.x.", "m.x."}, expected: []RangeEntry[string]{{"f.x.", "t.x."}, {"z.x.", "x."}}},
		{in: RangeEntry[string]{"f.x.", "n.x."}, expected: []RangeEntry[string]{{"f.x.", "t.x."}, {"z.x.", "x."}}},

		// entirely after
		{in: RangeEntry[string]{"u.x.", "v.x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"m.x.", "t.x."}, {"u.x.", "v.x."}, {"z.x.", "x."}}},

		// wrapping stuff
		// starts in first, wraps
		{in: RangeEntry[string]{"a.x.", "x."}, expected: []RangeEntry[string]{{"a.x.", "x."}}},

		{in: RangeEntry[string]{"f.x.", "x."}, expected: []RangeEntry[string]{{"f.x.", "x."}}},
		{in: RangeEntry[string]{"g.x.", "x."}, expected: []RangeEntry[string]{{"f.x.", "x."}}},
		{in: RangeEntry[string]{"h.x.", "x."}, expected: []RangeEntry[string]{{"f.x.", "x."}}},

		// starts between first and second, wraps
		{in: RangeEntry[string]{"i.x.", "x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"i.x.", "x."}}},

		// starts in second, wraps
		{in: RangeEntry[string]{"m.x.", "x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"m.x.", "x."}}},
		{in: RangeEntry[string]{"n.x.", "x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"m.x.", "x."}}},
		{in: RangeEntry[string]{"t.x.", "x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"m.x.", "x."}}},

		// starts after, wraps
		{in: RangeEntry[string]{"v.x.", "x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"m.x.", "t.x."}, {"v.x.", "x."}}},

		// no-op merge with wrapping range
		{in: RangeEntry[string]{"z.x.", "{.x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"m.x.", "t.x."}, {"z.x.", "x."}}},
		{in: RangeEntry[string]{"{.x.", "}.x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"m.x.", "t.x."}, {"z.x.", "x."}}},
		{in: RangeEntry[string]{"z.x.", "x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"m.x.", "t.x."}, {"z.x.", "x."}}},
		{in: RangeEntry[string]{"{.x.", "x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"m.x.", "t.x."}, {"z.x.", "x."}}},

		// found in the wild
		{start: []RangeEntry[string]{{"x.", "w.x."}, {"xr000.x.", "x."}}, in: RangeEntry[string]{"w.x.", "x."}, expected: []RangeEntry[string]{{"x.", "x."}}},
		{start: []RangeEntry[string]{{"x.", "2ajt2.x."}}, in: RangeEntry[string]{"4m4m4.x.", "x."}, expected: []RangeEntry[string]{{"x.", "2ajt2.x."}, {"4m4m4.x.", "x."}}},
		{start: []RangeEntry[string]{{"pl.x.", "po.x."}, {"sky.x.", "slap.x."}}, in: RangeEntry[string]{"slap.x.", "slat.x."}, expected: []RangeEntry[string]{{"pl.x.", "po.x."}, {"sky.x.", "slat.x."}}},
	} {
		start := defaultStart
		if datum.start != nil {
			start = datum.start
		}
		r := RangeSet[string]{Ranges: slices.Clone(start), Compare: dnsCompare, RWrapV: "x.", HasRWrap: true}
		r.Add(datum.in)
		if !slices.Equal(r.Ranges, datum.expected) {
			t.Errorf("initial data: %v, input: %v, expected: %v, actual: %v", start, datum.in, datum.expected, r.Ranges)
			return
		}
	}
}

func TestContains(t *testing.T) {
	for _, datum := range []struct {
		start    []RangeEntry[string]
		val      string
		expected bool
	}{
		// before first
		{val: "a.x.", expected: false},
		{val: "b.x.", expected: false},
		{val: "c.x.", expected: false},
		// in first
		{val: "f.x.", expected: true},
		{val: "g.x.", expected: true},
		// at end of range
		{val: "h.x.", expected: false},
		{val: "t.x.", expected: false},
		// in second
		{val: "m.x.", expected: true},
		{val: "n.x.", expected: true},
		{val: "s.x.", expected: true},
		// between first/second
		{val: "i.x.", expected: false},
		{val: "j.x.", expected: false},
		{val: "k.x.", expected: false},
		// between second/third
		{val: "u.x.", expected: false},
		{val: "v.x.", expected: false},
		{val: "w.x.", expected: false},
		// in rwrap range
		{val: "zz.x.", expected: true},
		{val: "zoo.x.", expected: true},
	} {
		start := defaultStart
		if datum.start != nil {
			start = datum.start
		}
		r := RangeSet[string]{Ranges: slices.Clone(start), Compare: dnsCompare, RWrapV: "x.", HasRWrap: true}
		if ret := r.Contains(datum.val); ret != datum.expected {
			t.Errorf("initial data: %v, val: %v, expected: %v, actual: %v", start, datum.val, datum.expected, ret)
			return
		}
	}
}

func TestContainsRange(t *testing.T) {
	for _, datum := range []struct {
		start    []RangeEntry[string]
		val      RangeEntry[string]
		expected bool
	}{
		// before first
		{val: RangeEntry[string]{Start: "a.x.", End: "a.x."}, expected: false},
		{val: RangeEntry[string]{Start: "a.x.", End: "b.x."}, expected: false},
		{val: RangeEntry[string]{Start: "b.x.", End: "d.x."}, expected: false},
		// before/in first
		{val: RangeEntry[string]{Start: "a.x.", End: "f.x."}, expected: false},
		{val: RangeEntry[string]{Start: "b.x.", End: "g.x."}, expected: false},
		{val: RangeEntry[string]{Start: "c.x.", End: "h.x."}, expected: false},
		// in/after first
		{val: RangeEntry[string]{Start: "g.x.", End: "i.x."}, expected: false},
		{val: RangeEntry[string]{Start: "f.x.", End: "j.x."}, expected: false},
		// in first
		{val: RangeEntry[string]{Start: "f.x.", End: "f.x."}, expected: true},
		{val: RangeEntry[string]{Start: "f.x.", End: "g.x."}, expected: true},
		{val: RangeEntry[string]{Start: "g.x.", End: "g.x."}, expected: true},
		{val: RangeEntry[string]{Start: "g.x.", End: "h.x."}, expected: true},
		// exact match on end of range
		{val: RangeEntry[string]{Start: "h.x.", End: "h.x."}, expected: false},
		{val: RangeEntry[string]{Start: "t.x.", End: "t.x."}, expected: false},
		// full range match
		{val: RangeEntry[string]{Start: "f.x.", End: "h.x."}, expected: true},
		{val: RangeEntry[string]{Start: "m.x.", End: "t.x."}, expected: true},
		// between ranges
		{val: RangeEntry[string]{Start: "i.x.", End: "j.x."}, expected: false},
		{val: RangeEntry[string]{Start: "w.x.", End: "x.x."}, expected: false},
		// bridges ranges
		{val: RangeEntry[string]{Start: "g.x.", End: "n.x."}, expected: false},
		{val: RangeEntry[string]{Start: "n.x.", End: "zz.x."}, expected: false},
		// in rwrap range
		{val: RangeEntry[string]{Start: "z.x.", End: "zz.x."}, expected: true},
		{val: RangeEntry[string]{Start: "zz.x.", End: "zzz.x."}, expected: true},
		// in rwrap range, ends with rwrap value
		{val: RangeEntry[string]{Start: "zzz.x.", End: "x."}, expected: true},
		// not in rwrap range, ends with rwrap value
		{val: RangeEntry[string]{Start: "w.x.", End: "x."}, expected: false},
		// in the wild
		{start: []RangeEntry[string]{{Start: "x.", End: "2ajt2.x."}}, val: RangeEntry[string]{Start: "4m4m4.x.", End: "x."}, expected: false},
	} {
		start := defaultStart
		if datum.start != nil {
			start = datum.start
		}
		r := RangeSet[string]{Ranges: slices.Clone(start), Compare: dnsCompare, RWrapV: "x.", HasRWrap: true}
		if ret := r.ContainsRange(datum.val); ret != datum.expected {
			t.Errorf("initial data: %v, val: %v, expected: %v, actual: %v", start, datum.val, datum.expected, ret)
			return
		}
	}
}

/*
func TestRemoveRanges(t *testing.T) {
	for _, datum := range []struct {
		start    []RangeEntry[string]
		in       RangeEntry[string]
		expected []RangeEntry[string]
	}{
		// takes off left end of single range
		{in: RangeEntry[string]{"a.x.", "g.x."}, expected: []RangeEntry[string]{{"g.x.", "h.x."}, {"m.x.", "t.x."}, {"z.x.", "x."}}},
		{in: RangeEntry[string]{"i.x.", "n.x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"n.x.", "t.x."}, {"z.x.", "x."}}},
		// takes off right end of single range
		{in: RangeEntry[string]{"g.x.", "i.x."}, expected: []RangeEntry[string]{{"f.x.", "g.x."}, {"m.x.", "t.x."}, {"z.x.", "x."}}},
		// splits a range
		{in: RangeEntry[string]{"n.x.", "o.x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"m.x.", "n.x."}, {"o.x.", "t.x."}, {"z.x.", "x."}}},
		// nop
		{in: RangeEntry[string]{"i.x.", "j.x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"m.x.", "t.x."}, {"z.x.", "x."}}},
		// takes off left and right end of two ranges, nothing inbetween
		{in: RangeEntry[string]{"g.x.", "n.x."}, expected: []RangeEntry[string]{{"f.x.", "g.x."}, {"n.x.", "t.x."}, {"z.x.", "x."}}},
		// takes off left and right end of two ranges, range inbetween
		{in: RangeEntry[string]{"g.x.", "zz.x."}, expected: []RangeEntry[string]{{"f.x.", "g.x."}, {"zz.x.", "x."}}},

		// exactly covers first range
		{in: RangeEntry[string]{"f.x.", "h.x."}, expected: []RangeEntry[string]{{"m.x.", "t.x."}, {"z.x.", "x."}}},
		// exactly covers other range
		{in: RangeEntry[string]{"m.x.", "t.x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"z.x.", "x."}}},
		// covers whole range +
		//{in: RangeEntry[string]{"m.x.", "u.x."}, expected: []RangeEntry[string]{{"g.x.", "h.x."}, {"z.x.", "x."}}},
		//{in: RangeEntry[string]{"i.x.", "u.x."}, expected: []RangeEntry[string]{{"g.x.", "h.x."}, {"z.x.", "x."}}},
		//{in: RangeEntry[string]{"i.x.", "t.x."}, expected: []RangeEntry[string]{{"g.x.", "h.x."}, {"z.x.", "x."}}},

		// splits range with wraparound
		{in: RangeEntry[string]{"zz.x.", "zzz.x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"m.x.", "t.x."}, {"z.x.", "zz.x."}, {"zzz.x.", "x."}}},
		// left end of wraparound range
		{in: RangeEntry[string]{"x.z.", "zz.x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"m.x.", "t.x."}, {"zz.x.", "x."}}},
		// right end of wraparound range
		{in: RangeEntry[string]{"zz.x.", "x."}, expected: []RangeEntry[string]{{"f.x.", "h.x."}, {"m.x.", "t.x."}, {"z.x.", "zz.x."}}},
	} {
		start := defaultStart
		if datum.start != nil {
			start = datum.start
		}
		r := RangeSet[string]{Ranges: slices.Clone(start), Compare: dnsCompare, RWrapV: "x.", HasRWrap: true}
		r.Remove(datum.in)
		if !slices.Equal(r.Ranges, datum.expected) {
			t.Errorf("initial data: %v, input: %v, expected: %v, actual: %v", start, datum.in, datum.expected, r.Ranges)
			return
		}
	}
}
*/

// everything below is taken from github.com/monoidic/dns (fork of miekg/dns) for testing purposes
// pasted in to avoid the dependency

func dnsCompare(s1, s2 string) int {
	s1b := doDDD([]byte(s1))
	s2b := doDDD([]byte(s2))

	s1 = string(s1b)
	s2 = string(s2b)

	s1lend := len(s1)
	s2lend := len(s2)

	for i := 0; ; i++ {
		s1lstart, end1 := PrevLabel(s1, i)
		s2lstart, end2 := PrevLabel(s2, i)

		if end1 && end2 {
			return 0
		}

		s1l := string(s1b[s1lstart:s1lend])
		s2l := string(s2b[s2lstart:s2lend])

		if cmp := labelCompare(s1l, s2l); cmp != 0 {
			return cmp
		}

		s1lend = s1lstart - 1
		s2lend = s2lstart - 1
		if s1lend == -1 {
			s1lend = 0
		}
		if s2lend == -1 {
			s2lend = 0
		}
	}
}

func doDDD(b []byte) []byte {
	lb := len(b)
	for i := 0; i < lb; i++ {
		if i+3 < lb && b[i] == '\\' && isDigit(b[i+1]) && isDigit(b[i+2]) && isDigit(b[i+3]) {
			b[i] = dddToByte(b[i+1 : i+4])
			for j := i + 1; j < lb-3; j++ {
				b[j] = b[j+3]
			}
			lb -= 3
		}
	}
	return b[:lb]
}

func isDigit(b byte) bool { return b >= '0' && b <= '9' }

func dddToByte(s []byte) byte {
	_ = s[2] // bounds check hint to compiler; see golang.org/issue/14808
	return byte((s[0]-'0')*100 + (s[1]-'0')*10 + (s[2] - '0'))
}

func PrevLabel(s string, n int) (i int, start bool) {
	if s == "" {
		return 0, true
	}
	if n == 0 {
		return len(s), false
	}

	l := len(s) - 1
	if s[l] == '.' {
		l--
	}

	for ; l >= 0 && n > 0; l-- {
		if s[l] != '.' {
			continue
		}
		j := l - 1
		for j >= 0 && s[j] == '\\' {
			j--
		}

		if (j-l)%2 == 0 {
			continue
		}

		n--
		if n == 0 {
			return l + 1, false
		}
	}

	return 0, n > 1
}

func labelCompare(a, b string) int {
	la := len(a)
	lb := len(b)
	minLen := la
	if lb < la {
		minLen = lb
	}
	for i := 0; i < minLen; i++ {
		ai := a[i]
		bi := b[i]
		if ai >= 'A' && ai <= 'Z' {
			ai |= 'a' - 'A'
		}
		if bi >= 'A' && bi <= 'Z' {
			bi |= 'a' - 'A'
		}
		if ai != bi {
			if ai > bi {
				return 1
			}
			return -1
		}
	}

	if la > lb {
		return 1
	} else if la < lb {
		return -1
	}
	return 0
}
