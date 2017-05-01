package types

import (
	"github.com/jinzhu/copier"
)

// Firewall defines a collection of firewall rules
type Firewall []FirewallRule

// ToMap returns a slice of map[string]string version
func (f Firewall) ToMap() []map[string]string {
	var sm []map[string]string
	for _, m := range f {
		sm = append(sm, map[string]string{
			"destination":     m.Destination,
			"destinationPort": m.DestinationPort,
			"protocol":        m.Protocol,
		})
	}
	return sm
}

// Eql checks whether another firewall object is equal.
func (f Firewall) Eql(o Firewall) bool {
	l := len(f)
	found := 0
	if l != len(o) {
		return false
	}
	for i := 0; i < l; i++ {
		cur := f[i]
		for j := 0; j < l; j++ {
			oCur := o[j]
			if cur.Eql(oCur) {
				found++
			}
		}
	}
	return found == l
}

// DeDup removes duplicate firewall rules
func (f Firewall) DeDup() Firewall {
	newFirewall := Firewall{}
	for _, x := range f {
		isEql := false
		for _, y := range newFirewall {
			if x.Eql(y) {
				isEql = true
				continue
			}
		}
		if !isEql {
			newFirewall = append(newFirewall, x)
		}
	}
	return newFirewall
}

// FirewallRule represents information about a destination to allow connections to.
type FirewallRule struct {
	Destination     string `structs:"destination" mapstructure:"destination,omitempty"`
	DestinationPort string `structs:"destinationPort" mapstructure:"destinationPort,omitempty"`
	Protocol        string `structs:"protocol" mapstructure:"protocol,omitempty"`
}

// Eql returns true if another firewall rule is equal
func (r FirewallRule) Eql(o FirewallRule) bool {
	return r.Destination == o.Destination && r.DestinationPort == o.DestinationPort && r.Protocol == o.Protocol
}

// CopyFirewall  takes a object that is similar to a Firewall object,
// copies the items/fields that are similar into a new Firewall object.
// This method does not return any error.
func CopyFirewall(o interface{}) Firewall {
	var fw Firewall
	copier.Copy(&fw, o)
	return fw
}
