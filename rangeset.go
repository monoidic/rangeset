package rangeset

import (
	"slices"
	"sort"
)

type RangeEntry[T any] struct {
	Start T
	End   T
}

// container for arbitrary ranges of values
type RangeSet[T any] struct {
	// the merged ranges
	Ranges []RangeEntry[T]
	// a three-way comparison function like strcmp;
	// 0 for equality, -1 for v1 < v2, 1 for v1 > v2
	Compare func(v1, v2 T) int
	// a sentinel value indicating wrapping around from this value from the end to the start
	// if HasWrap is true and the final value is this, then any value sorting after it
	// is considered within the range
	RWrapV T
	// whether or not there is a "wraparound value" on the right side
	HasRWrap bool
}

// helper to check whether a given value is in one of the ranges + return the index of the range
func (r *RangeSet[T]) containsI(v T) (int, bool) {
	l := len(r.Ranges)

	// empty set
	if l == 0 {
		return 0, false
	}

	// whether or not the end of the last range is the wrap value
	endWraps := r.HasRWrap && r.Compare(r.Ranges[l-1].End, r.RWrapV) == 0

	// value is in the wrapped area
	if endWraps && r.Compare(r.Ranges[l-1].Start, v) != 1 {
		return l - 1, true
	}

	i := sort.Search(l, func(i int) bool {
		if endWraps && i == l-1 {
			return true
		}
		return r.Compare(v, r.Ranges[i].End) == -1
	})

	if i == l {
		return 0, false
	}

	rn := r.Ranges[i]
	start, end := rn.Start, rn.End
	// value is within the range
	return i, r.Compare(v, end) == -1 && r.Compare(start, v) != 1
}

// check whether a given value is contained within the range set
func (r *RangeSet[T]) Contains(v T) bool {
	_, ret := r.containsI(v)
	return ret
}

// add a range, potentially expanding or merging existing ranges
func (r *RangeSet[T]) Add(newEntry RangeEntry[T]) {
	if len(r.Ranges) == 0 {
		// first range
		r.Ranges = []RangeEntry[T]{newEntry}
		return
	}

	// whether or not the end of the last range is the wrap value
	endWraps := r.HasRWrap && r.Compare(r.Ranges[len(r.Ranges)-1].End, r.RWrapV) == 0

	startI := r.addStart(&newEntry, endWraps)
	endI := r.addEnd(&newEntry, endWraps)

	// remove (possibly empty) range of values which will be merged
	r.Ranges = slices.Delete(r.Ranges, startI, endI)
	// insert (possibly merged from existing removed ranges) range
	r.Ranges = slices.Insert(r.Ranges, startI, newEntry)
}

func (r *RangeSet[T]) addStart(newEntry *RangeEntry[T], endWraps bool) int {
	l := len(r.Ranges)
	startI := sort.Search(l, func(i int) bool {
		if endWraps && i == l-1 {
			return true
		}
		return r.Compare(newEntry.Start, r.Ranges[i].Start) == -1
	})

	if startI == l {
		if r.Compare(newEntry.End, r.Ranges[l-1].End) == -1 {
			// is entirely after
			return startI
		}
		// still interacts with the last range
		startI--
	}

	switch r.Compare(newEntry.Start, r.Ranges[startI].Start) {
	case -1:
		// expand left to previous range?
		if startI == 0 {
			// cannot expand left
			break
		}
		if r.Compare(newEntry.Start, r.Ranges[startI-1].End) != 1 {
			// merge left
			startI--
			newEntry.Start = r.Ranges[startI].Start
		}
	case 0:
		// ranges start at same spot, nop
	case 1:
		// expand within range
		newEntry.Start = r.Ranges[startI].Start
	}

	return startI
}

func (r *RangeSet[T]) addEnd(newEntry *RangeEntry[T], endWraps bool) int {
	l := len(r.Ranges)

	if r.HasRWrap && r.Compare(newEntry.End, r.RWrapV) == 0 {
		return l
	}

	endI := sort.Search(l, func(i int) bool {
		if endWraps && i == l-1 {
			return true
		}
		return r.Compare(newEntry.End, r.Ranges[i].End) != 1
	})

	if endI != l && r.Compare(r.Ranges[endI].Start, newEntry.End) != 1 {
		// connects ranges, simply merge
		newEntry.End = r.Ranges[endI].End
		endI++
	}

	return endI
}

func (r *RangeSet[T]) ContainsRange(rn RangeEntry[T]) bool {
	// a range is contained entirely if both the start and end exist
	// and are contained within the same defined range

	startI, startMatch := r.containsI(rn.Start)
	if !startMatch {
		return false
	}

	l := len(r.Ranges)
	endWraps := r.HasRWrap && r.Compare(r.Ranges[l-1].End, r.RWrapV) == 0
	if endWraps && (r.Compare(r.Ranges[l-1].Start, rn.End) == -1 || r.Compare(rn.End, r.RWrapV) == 0) {
		return startI == l-1
	}

	endI := sort.Search(l, func(i int) bool {
		if endWraps && i == l-1 {
			return true
		}
		return r.Compare(rn.End, r.Ranges[i].End) != 1
	})

	if startI != endI {
		return false
	}

	return r.Compare(r.Ranges[endI].Start, rn.End) != 1 && r.Compare(rn.End, r.Ranges[endI].End) != 1
}
