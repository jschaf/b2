package track

// Code is a tracking code embedded in the hash of the URL like "z_qux". Used
// to figure out where traffic comes from.
type Code string

// Referrer is the source of traffic for a Code.
type Referrer string

const (
	Point72VC Referrer = "Point 72 VC"
)

// AllRefs returns the mapping between a unique code the Referrer.
func AllRefs() map[Code]Referrer {
	return map[Code]Referrer{
		"z_mead": Point72VC,
	}
}
