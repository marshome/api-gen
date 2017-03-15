package generate

//A Param or Property.
type Field interface {
	Default() string
	Enum() []string
	EnumDescriptions() []string
	UnfortunateDefault() bool
}
