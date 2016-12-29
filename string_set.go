// Generated by: main
// TypeWriter: set
// Directive: +gen on String

package main

// Set is a modification of https://github.com/deckarep/golang-set
// The MIT License (MIT)
// Copyright (c) 2013 Ralph Caraveo (deckarep@gmail.com)

// StringSet is the primary type that represents a set
type StringSet map[String]struct{}

// NewStringSet creates and returns a reference to an empty set.
func NewStringSet(a ...String) StringSet {
	s := make(StringSet)
	for _, i := range a {
		s.Add(i)
	}
	return s
}

// ToSlice returns the elements of the current set as a slice
func (set StringSet) ToSlice() []String {
	var s []String
	for v := range set {
		s = append(s, v)
	}
	return s
}

// Add adds an item to the current set if it doesn't already exist in the set.
func (set StringSet) Add(i String) bool {
	_, found := set[i]
	set[i] = struct{}{}
	return !found //False if it existed already
}

// Contains determines if a given item is already in the set.
func (set StringSet) Contains(i String) bool {
	_, found := set[i]
	return found
}

// ContainsAll determines if the given items are all in the set
func (set StringSet) ContainsAll(i ...String) bool {
	for _, v := range i {
		if !set.Contains(v) {
			return false
		}
	}
	return true
}

// IsSubset determines if every item in the other set is in this set.
func (set StringSet) IsSubset(other StringSet) bool {
	for elem := range set {
		if !other.Contains(elem) {
			return false
		}
	}
	return true
}

// IsSuperset determines if every item of this set is in the other set.
func (set StringSet) IsSuperset(other StringSet) bool {
	return other.IsSubset(set)
}

// Union returns a new set with all items in both sets.
func (set StringSet) Union(other StringSet) StringSet {
	unionedSet := NewStringSet()

	for elem := range set {
		unionedSet.Add(elem)
	}
	for elem := range other {
		unionedSet.Add(elem)
	}
	return unionedSet
}

// Intersect returns a new set with items that exist only in both sets.
func (set StringSet) Intersect(other StringSet) StringSet {
	intersection := NewStringSet()
	// loop over smaller set
	if set.Cardinality() < other.Cardinality() {
		for elem := range set {
			if other.Contains(elem) {
				intersection.Add(elem)
			}
		}
	} else {
		for elem := range other {
			if set.Contains(elem) {
				intersection.Add(elem)
			}
		}
	}
	return intersection
}

// Difference returns a new set with items in the current set but not in the other set
func (set StringSet) Difference(other StringSet) StringSet {
	differencedSet := NewStringSet()
	for elem := range set {
		if !other.Contains(elem) {
			differencedSet.Add(elem)
		}
	}
	return differencedSet
}

// SymmetricDifference returns a new set with items in the current set or the other set but not in both.
func (set StringSet) SymmetricDifference(other StringSet) StringSet {
	aDiff := set.Difference(other)
	bDiff := other.Difference(set)
	return aDiff.Union(bDiff)
}

// Clear clears the entire set to be the empty set.
func (set *StringSet) Clear() {
	*set = make(StringSet)
}

// Remove allows the removal of a single item in the set.
func (set StringSet) Remove(i String) {
	delete(set, i)
}

// Cardinality returns how many items are currently in the set.
func (set StringSet) Cardinality() int {
	return len(set)
}

// Iter returns a channel of type String that you can range over.
func (set StringSet) Iter() <-chan String {
	ch := make(chan String)
	go func() {
		for elem := range set {
			ch <- elem
		}
		close(ch)
	}()

	return ch
}

// Equal determines if two sets are equal to each other.
// If they both are the same size and have the same items they are considered equal.
// Order of items is not relevent for sets to be equal.
func (set StringSet) Equal(other StringSet) bool {
	if set.Cardinality() != other.Cardinality() {
		return false
	}
	for elem := range set {
		if !other.Contains(elem) {
			return false
		}
	}
	return true
}

// Clone returns a clone of the set.
// Does NOT clone the underlying elements.
func (set StringSet) Clone() StringSet {
	clonedSet := NewStringSet()
	for elem := range set {
		clonedSet.Add(elem)
	}
	return clonedSet
}
