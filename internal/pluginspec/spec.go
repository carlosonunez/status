// Package pluginspec defines shared types used to describe EventGetter and
// StatusSetter parameters. It has no dependencies on other internal packages
// so both getter and setter can import it without creating a cycle.
package pluginspec

// Type constants describe the expected value type for a ParamSpec.
const (
	TypeString = "string"
	TypeBool   = "bool"
	TypeInt    = "int"
)

// ParamSpec describes a single parameter accepted by an EventGetter or
// StatusSetter. It is used by the "status integration list/show" commands.
type ParamSpec struct {
	Name        string
	Description string
	Required    bool
	Type        string // one of the Type* constants
}
