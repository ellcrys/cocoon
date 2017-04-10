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
	} else if len(targetParts) == 2 { // [ledgerName, actorID]
		if m[targetParts[0]] != nil { // ledger name already exists and has value
			if _, ok := m[targetParts[0]].(string); ok { // has string value
				m[targetParts[0]] = map[string]string{ // replace with new map structure
					targetParts[1]: privileges,
				}
			}
			if innerMap, ok := m[targetParts[0]].(map[string]string); ok { // hash map value
				innerMap[targetParts[1]] = privileges // update existing map value
			}
			if innerMap, ok := m[targetParts[0]].(map[string]interface{}); ok { // hash map value
				innerMap[targetParts[1]] = privileges // update existing map value
			}
		} else { // does not exist, free to add map value
			m[targetParts[0]] = map[string]string{
				targetParts[1]: privileges,
			}
		}
	} else {
		return fmt.Errorf("target format is invalid")
	}
	return nil
}

// Remove removes a target and its value
func (m ACLMap) Remove(target string) {
	if m == nil {
		return
	}
	targetParts := strings.Split(target, ".")
	if len(targetParts) == 1 {
		delete(m, targetParts[0])
	} else {
		if m[targetParts[0]] != nil {
			if innerMap, ok := m[targetParts[0]].(map[string]interface{}); ok {
				delete(innerMap, targetParts[1])
				if len(innerMap) == 0 { // if target is now empty, just remove it entirely
					delete(m, targetParts[0])
				}
			}
			if innerMap, ok := m[targetParts[0]].(map[string]string); ok {
				delete(innerMap, targetParts[1])
				if len(innerMap) == 0 { // if target is now empty, just remove it entirely
					delete(m, targetParts[0])
				}
			}
		}
	}
}
