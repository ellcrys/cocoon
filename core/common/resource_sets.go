package common

// SupportedResourceSets defines the support sets of resources
// that can be allocated to a task
var SupportedResourceSets = map[string]map[string]int{
	"s1": {
		"cpuShare": 100,
		"memory":   256,
		"disk":     4000,
	},
	"s2": {
		"cpuShare": 100,
		"memory":   512,
		"disk":     4000,
	},
	"m1": {
		"cpuShare": 100,
		"memory":   1024,
		"disk":     4000,
	},
	"m2": {
		"cpuShare": 200,
		"memory":   2048,
		"disk":     4000,
	},
}

// GetResourceSet returns the full resource where the resource
// attributes match the one passed
func GetResourceSet(memory, cpuShare int) map[string]map[string]int {
	for n, set := range SupportedResourceSets {
		if set["memory"] == memory && set["cpuShare"] == cpuShare {
			return map[string]map[string]int{
				n: set,
			}
		}
	}
	return nil
}
