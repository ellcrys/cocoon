package mapdiff

import "reflect"

// DiffType represents a difference type
type DiffType string

// DiffTypeMissing is a difference type that represents a missing field
const DiffTypeMissing DiffType = "missing"

// DiffTypeDifferent represents a difference type that represents a difference between two fields
const DiffTypeDifferent DiffType = "different"

// MapDiff defines a structure that performs
// shallow difference checks between one main map and other maps
type MapDiff struct {
	main        Map
	againstMaps []Map
}

// DiffValue represents a difference
type DiffValue struct {
	AffectedField []interface{}
	Type          DiffType
	LongestMap    Map
	AffectedMap   Map
}

// Map defines a structure for representing the maps to check
type Map struct {
	Value map[string]interface{}
	Name  string
}

// NewMapDiff creates an new instance of MapDiff to
// perform difference checks between a map type and one or more others
func NewMapDiff(m Map, against ...Map) *MapDiff {
	return &MapDiff{
		main:        m,
		againstMaps: against,
	}
}

// Difference checks whether the main map is different from the other maps
func (s *MapDiff) Difference() []bool {
	result := make([]bool, len(s.againstMaps))
	for i, m := range s.againstMaps {
		result[i] = !reflect.DeepEqual(s.main.Value, m.Value)
	}
	return result
}

func (s *MapDiff) diff(a, b Map) []DiffValue {
	var longest = a.Value
	var shortest = b.Value
	var lMap = a
	var sMap = b
	if len(b.Value) > len(a.Value) {
		longest = b.Value
		shortest = a.Value
		sMap = a
		lMap = b
	}

	result := []DiffValue{}
	for aKey, aVal := range longest {
		isMissingKey := true
		diffVals := []interface{}{}
		for bKey, bVal := range shortest {
			if aKey == bKey {
				isMissingKey = false
				if !reflect.DeepEqual(aVal, bVal) {
					diffVals = append(diffVals, map[string]interface{}{aKey: aVal})
				}
				continue
			}
		}
		if isMissingKey {
			result = append(result, DiffValue{
				LongestMap: lMap,
				Type:       DiffTypeMissing,
				AffectedField: []interface{}{
					map[string]interface{}{aKey: aVal},
				},
				AffectedMap: sMap,
			})
		}
		if len(diffVals) > 0 {
			result = append(result, DiffValue{
				LongestMap:    lMap,
				Type:          DiffTypeDifferent,
				AffectedField: diffVals,
				AffectedMap:   sMap,
			})
		}
	}
	return result
}

// Diff checks the main map against the other map(s) and
// returns the difference for each map. If there is no difference
// between the main map and a map being compared, nil is included in the
// corresponding index of the checked map.
func (s *MapDiff) Diff() [][]DiffValue {
	var result = [][]DiffValue{}
	for _, m := range s.againstMaps {
		if diffVals := s.diff(s.main, m); len(diffVals) > 0 {
			result = append(result, diffVals)
		} else {
			result = append(result, nil)
		}
	}
	return result
}
