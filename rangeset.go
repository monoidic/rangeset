package rangeset

import (
	"errors"
	"iter"
	"sort"

	"github.com/tidwall/btree"
)

type RangeEntry[T any] struct {
	Start T
	End   T
}

// container for arbitrary ranges of values
type RangeSet[T any] struct {
	// the merged ranges
	Ranges *btree.BTreeG[RangeEntry[T]]
	// a three-way comparison function like strcmp;
	// 0 for equality, -1 for v1 < v2, 1 for v1 > v2
	Compare func(v1, v2 T) int
	// a sentinel value indicating wrapping around from this value from the end to the start
	// if HasWrap is true and the final value is this, then any value sorting after it
	// is considered within the range
	RWrapV T
	// whether or not there is a "wraparound value" on the right side
	HasRWrap bool
	// maybe speeds stuff up idk
	hint btree.PathHint
}

var ErrIndex = errors.New("index out of range of RangeSet")

func NewRangeset[T any](compare func(v1, v2 T) int, rwrapV T, hasWrap bool) *RangeSet[T] {
	less := func(v1, v2 RangeEntry[T]) bool { return compare(v1.Start, v2.Start) == -1 }
	return &RangeSet[T]{
		// do not use locks; already handled externally anyhow
		Ranges:   btree.NewBTreeGOptions(less, btree.Options{NoLocks: true}),
		RWrapV:   rwrapV,
		HasRWrap: hasWrap,
		Compare:  compare,
	}
}

// helper to check whether a given value is in one of the ranges + return the index of the range
func (r *RangeSet[T]) containsI(v T) (int, bool) {
	l := r.Ranges.Len()

	// empty set
	if l == 0 {
		return 0, false
	}

	last := r.checkedGetAt(l - 1)

	// whether or not the end of the last range is the wrap value
	endWraps := r.HasRWrap && r.Compare(last.End, r.RWrapV) == 0

	// value is in the wrapped area
	if endWraps && r.Compare(last.Start, v) != 1 {
		return l - 1, true
	}

	i := sort.Search(l, func(i int) bool {
		if endWraps && i == l-1 {
			return true
		}
		return r.Compare(v, r.checkedGetAt(i).End) == -1
	})

	if i == l {
		return 0, false
	}

	rn := r.checkedGetAt(i)
	start, end := rn.Start, rn.End
	// value is within the range
	// start <= v < end
	return i, r.Compare(start, v) != 1 && r.Compare(v, end) == -1
}

func (r *RangeSet[T]) checkedGetAt(i int) RangeEntry[T] {
	ret, ok := r.Ranges.GetAt(i)
	if !ok {
		panic("btree is broken")
	}
	return ret
}

// check whether a given value is contained within the range set
func (r *RangeSet[T]) Contains(v T) bool {
	_, ret := r.containsI(v)
	return ret
}

// add a range, potentially expanding or merging existing ranges
func (r *RangeSet[T]) Add(newEntry RangeEntry[T]) {
	l := r.Ranges.Len()
	if r.Ranges.Len() == 0 {
		// first range
		r.Ranges.SetHint(newEntry, &r.hint)
		return
	}

	// whether or not the end of the last range is the wrap value
	last := r.checkedGetAt(l - 1)
	endWraps := r.HasRWrap && r.Compare(last.End, r.RWrapV) == 0

	startI := r.addStart(&newEntry, endWraps)
	endI := r.addEnd(&newEntry, endWraps)

	// remove (possibly empty) range of values which will be merged

	opts := &btree.DeleteRangeOptions{NoReturn: true}

	if endI == l {
		opts.MaxInclusive = true
		endI--
	}

	if (startI < endI) || (startI == endI && opts.MaxInclusive) {
		min := r.checkedGetAt(startI)
		var max RangeEntry[T]
		if startI == endI {
			max = min
		} else {
			max = r.checkedGetAt(endI)
		}
		r.Ranges.DeleteRange(min, max, opts)
	}
	// insert (possibly merged from existing removed ranges) range
	r.Ranges.SetHint(newEntry, &r.hint)
}

func (r *RangeSet[T]) addStart(newEntry *RangeEntry[T], endWraps bool) int {
	l := r.Ranges.Len()
	startI := sort.Search(l, func(i int) bool {
		if endWraps && i == l-1 {
			return true
		}
		return r.Compare(newEntry.Start, r.checkedGetAt(i).Start) == -1
	})

	if startI == l {
		if r.Compare(r.checkedGetAt(l-1).End, newEntry.Start) == -1 {
			// range is entirely after last known range
			return startI
		}
		// still interacts with the last range
		startI--
	}

	startIv := r.checkedGetAt(startI)
	switch r.Compare(newEntry.Start, startIv.Start) {
	case -1:
		// expand left to previous range?
		if startI == 0 {
			// cannot expand left
			break
		}
		startIprevV := r.checkedGetAt(startI - 1)
		if r.Compare(newEntry.Start, startIprevV.End) != 1 {
			// merge left
			startI--
			newEntry.Start = startIprevV.Start
		}
	case 0:
		// ranges start at same spot, nop
	case 1:
		// expand within range
		newEntry.Start = startIv.Start
	}

	return startI
}

func (r *RangeSet[T]) addEnd(newEntry *RangeEntry[T], endWraps bool) int {
	l := r.Ranges.Len()

	if r.HasRWrap && r.Compare(newEntry.End, r.RWrapV) == 0 {
		return l
	}

	endI := sort.Search(l, func(i int) bool {
		if endWraps && i == l-1 {
			return true
		}
		return r.Compare(newEntry.End, r.checkedGetAt(i).End) != 1
	})

	if endI != l {
		endV := r.checkedGetAt(endI)
		if r.Compare(endV.Start, newEntry.End) != 1 {
			// connects ranges, simply merge
			newEntry.End = endV.End
			endI++
		}
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

	l := r.Ranges.Len()

	lastV := r.checkedGetAt(l - 1)
	endWraps := r.HasRWrap && r.Compare(lastV.End, r.RWrapV) == 0
	if endWraps && (r.Compare(lastV.Start, rn.End) == -1 || r.Compare(rn.End, r.RWrapV) == 0) {
		return startI == l-1
	}

	endI := sort.Search(l, func(i int) bool {
		if endWraps && i == l-1 {
			return true
		}
		return r.Compare(rn.End, r.checkedGetAt(i).End) != 1
	})

	if startI != endI {
		return false
	}

	endV := r.checkedGetAt(endI)
	// endV.start <= rn.end <= endV.end
	return r.Compare(endV.Start, rn.End) != 1 && r.Compare(rn.End, endV.End) != 1
}

func (r *RangeSet[T]) Items() []RangeEntry[T] {
	return r.Ranges.Items()
}

func (r *RangeSet[T]) Iter() iter.Seq[RangeEntry[T]] {
	return func(yield func(RangeEntry[T]) bool) {
		iterator := r.Ranges.Iter()
		defer iterator.Release()
		for iterator.Next() {
			if !yield(iterator.Item()) {
				return
			}
		}
	}
}

func (r *RangeSet[T]) Get(idx int) (RangeEntry[T], error) {
	var err error
	ret, ok := r.Ranges.GetAt(idx)
	if !ok {
		err = ErrIndex
	}
	return ret, err
}

func (r *RangeSet[T]) Len() int {
	return r.Ranges.Len()
}
