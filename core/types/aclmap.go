package types

import (
	"fmt"
	"strings"
)

// ACLMap represents an ACL rule collection
type ACLMap map[string]interface{}

// NewACLMap creates an ACLMap from a default value
func NewACLMap(defValue map[string]interface{}) ACLMap {
	aclMap := ACLMap(defValue)
	return aclMap
}

// Add a new entry.
func (m ACLMap) Add(target, privileges string) error {

	if m == nil {
		m = map[string]interface{}{}
	}

	targetParts := strings.Split(target, ".")
	if len(targetParts) == 1 {
		m[targetParts[0]] = privileges
	} else if len(targetParts) == 2 {
		if m[targetParts[0]] != nil {
			if _, ok := m[targetParts[0]].(string); ok {
				m[targetParts[0]] = map[string]string{
					targetParts[1]: privileges,
				}
			}
			if innerMap, ok := m[targetParts[0]].(map[string]interface{}); ok {
				innerMap[targetParts[1]] = privileges
			}
		} else {
			m[targetParts[0]] = map[string]string{
				targetParts[1]: privileges,
			}
		}
	} else {
		return fmt.Errorf("target format is invalid")
	}
	return nil
}
