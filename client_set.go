// Generated by: main
// TypeWriter: set
// Directive: +gen on *Client

package middleman

// Set is a modification of https://github.com/deckarep/golang-set
// The MIT License (MIT)
// Copyright (c) 2013 Ralph Caraveo (deckarep@gmail.com)

// ClientSet is the primary type that represents a set
type ClientSet map[*Client]struct{}

// NewClientSet creates and returns a reference to an empty set.
func NewClientSet(a ...*Client) ClientSet {
	s := make(ClientSet)
	for _, i := range a {
		s.Add(i)
	}
	return s
}

// ToSlice returns the elements of the current set as a slice
func (set ClientSet) ToSlice() []*Client {
	var s []*Client
	for v := range set {
		s = append(s, v)
	}
	return s
}

// Add adds an item to the current set if it doesn't already exist in the set.
func (set ClientSet) Add(i *Client) bool {
	_, found := set[i]
	set[i] = struct{}{}
	return !found //False if it existed already
}

// Contains determines if a given item is already in the set.
func (set ClientSet) Contains(i *Client) bool {
	_, found := set[i]
	return found
}

// ContainsAll determines if the given items are all in the set
func (set ClientSet) ContainsAll(i ...*Client) bool {
	for _, v := range i {
		if !set.Contains(v) {
			return false
		}
	}
	return true
}

// IsSubset determines if every item in the other set is in this set.
func (set ClientSet) IsSubset(other ClientSet) bool {
	for elem := range set {
		if !other.Contains(elem) {
			return false
		}
	}
	return true
}

// IsSuperset determines if every item of this set is in the other set.
func (set ClientSet) IsSuperset(other ClientSet) bool {
	return other.IsSubset(set)
}

// Union returns a new set with all items in both sets.
func (set ClientSet) Union(other ClientSet) ClientSet {
	unionedSet := NewClientSet()

	for elem := range set {
		unionedSet.Add(elem)
	}
	for elem := range other {
		unionedSet.Add(elem)
	}
	return unionedSet
}

// Intersect returns a new set with items that exist only in both sets.
func (set ClientSet) Intersect(other ClientSet) ClientSet {
	intersection := NewClientSet()
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
func (set ClientSet) Difference(other ClientSet) ClientSet {
	differencedSet := NewClientSet()
	for elem := range set {
		if !other.Contains(elem) {
			differencedSet.Add(elem)
		}
	}
	return differencedSet
}

// SymmetricDifference returns a new set with items in the current set or the other set but not in both.
func (set ClientSet) SymmetricDifference(other ClientSet) ClientSet {
	aDiff := set.Difference(other)
	bDiff := other.Difference(set)
	return aDiff.Union(bDiff)
}

// Clear clears the entire set to be the empty set.
func (set *ClientSet) Clear() {
	*set = make(ClientSet)
}

// Remove allows the removal of a single item in the set.
func (set ClientSet) Remove(i *Client) {
	delete(set, i)
}

// Cardinality returns how many items are currently in the set.
func (set ClientSet) Cardinality() int {
	return len(set)
}

// Iter returns a channel of type *Client that you can range over.
func (set ClientSet) Iter() <-chan *Client {
	ch := make(chan *Client)
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
func (set ClientSet) Equal(other ClientSet) bool {
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
func (set ClientSet) Clone() ClientSet {
	clonedSet := NewClientSet()
	for elem := range set {
		clonedSet.Add(elem)
	}
	return clonedSet
}
