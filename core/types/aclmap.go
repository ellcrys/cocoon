package types

import (
	"fmt"
	"strings"

	"reflect"

	"github.com/ellcrys/util"
)

// ACLMap represents an ACL rule collection
type ACLMap map[string]interface{}

// NewACLMap creates an ACLMap from a default value
func NewACLMap(defValue map[string]interface{}) ACLMap {
	aclMap := ACLMap(defValue)
	return aclMap
}

// NewACLMapFromByte takes a byte slice representing a json encoded ACL data
// and returns an ACLMap. If aclByte could not be coerced from JSON, an empty
// ACLMap is returned. Caller should ensure aclBytes is valid JSON.
func NewACLMapFromByte(aclByte []byte) ACLMap {
	var aclMap map[string]interface{}
	util.FromJSON(aclByte, &aclMap)
	return NewACLMap(aclMap)
}

// ToJSON returns a json encoded representation of the ACLMap
func (m ACLMap) ToJSON() []byte {
	b, _ := util.ToJSON(m)
	return b
}

// Eql checks whether another ACLMap is equal
func (m ACLMap) Eql(o ACLMap) bool {
	if len(m) != len(o) {
		return false
	}
	for k, v := range m {
		if o[k] == nil {
			return false
		}
		if !reflect.DeepEqual(v, o[k]) {
			return false
		}
	}
	return true
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
