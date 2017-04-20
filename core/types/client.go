package types

import (
	"fmt"

	"github.com/ellcrys/util"
)

// Cocoon represents a smart contract application
type Cocoon struct {
	IdentityID          string   `structs:"identity_id" mapstructure:"identity_id"`
	ID                  string   `structs:"ID" mapstructure:"ID"`
	URL                 string   `structs:"URL" mapstructure:"URL"`
	Version             string   `structs:"version" mapstructure:"version"`
	Language            string   `structs:"language" mapstructure:"language"`
	BuildParam          string   `structs:"buildParam" mapstructure:"buildParam"`
	Memory              int      `structs:"memory" mapstructure:"memory"`
	CPUShare            int      `structs:"CPUShare" mapstructure:"CPUShare"`
	NumSignatories      int      `structs:"numSignatories" mapstructure:"numSignatories"`
	SigThreshold        int      `structs:"sigThreshold" mapstructure:"sigThreshold"`
	Link                string   `structs:"link" mapstructure:"link"`
	Releases            []string `structs:"releases" mapstructure:"releases"`
	Signatories         []string `structs:"signatories" mapstructure:"signatories"`
	Status              string   `structs:"status" mapstructure:"status"`
	LastDeployedRelease string   `structs:"lastDeployedRelease" mapstructure:"lastDeployedRelease"`
	ACL                 ACLMap   `structs:"ACL" mapstructure:"ACL"`
	Firewall            Firewall `structs:"firewall" mapstructure:"firewall"`
	CreatedAt           string   `structs:"createdAt" mapstructure:"createdAt"`
}

// ToJSON returns the json equivalent of this object
func (c *Cocoon) ToJSON() []byte {
	json, _ := util.ToJSON(c)
	return json
}

// MakeCocoonKey constructs a cocoon key
func MakeCocoonKey(id string) string {
	return fmt.Sprintf("cocoon;%s", id)
}

// Firewall defines a collection of firewall rules
type Firewall []FirewallRule

// ToMap returns a map[string]string version
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
	Destination     string `structs:"destination" mapstructure:"destination"`
	DestinationPort string `structs:"destinationPort" mapstructure:"destinationPort"`
	Protocol        string `structs:"protocol" mapstructure:"protocol"`
}

// Eql returns true if another firewall rule is equal
func (r FirewallRule) Eql(o FirewallRule) bool {
	return r.Destination == o.Destination && r.DestinationPort == o.DestinationPort && r.Protocol == o.Protocol
}

// Release represents a new update to a cocoon's
// configuration
type Release struct {
	ID          string   `structs:"ID" mapstructure:"ID"`
	CocoonID    string   `structs:"cocoonID" mapstructure:"cocoonID"`
	URL         string   `structs:"URL" mapstructure:"URL"`
	Version     string   `structs:"version" mapstructure:"version"`
	Language    string   `structs:"language" mapstructure:"language"`
	BuildParam  string   `structs:"buildParam" mapstructure:"buildParam"`
	Link        string   `structs:"link" mapstructure:"link"`
	SigApproved int      `structs:"sigApproved" mapstructure:"sigApproved"`
	SigDenied   int      `structs:"sigDenied" mapstructure:"sigDenied"`
	VotersID    []string `structs:"votersID" mapstructure:"votersID"`
	Firewall    Firewall `structs:"firewall" mapstructure:"firewall"`
	ACL         ACLMap   `structs:"acl" mapstructure:"acl"`
	CreatedAt   string   `structs:"createdAt" mapstructure:"createdAt"`
}

// ToJSON returns the json equivalent of this object
func (r *Release) ToJSON() []byte {
	json, _ := util.ToJSON(r)
	return json
}

// MakeReleaseKey constructs an release key
func MakeReleaseKey(id string) string {
	return fmt.Sprintf("release;%s", id)
}
