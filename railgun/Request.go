package railgun

type Request interface {
	String() string
	Map() map[string]interface{}
	Serialize(format string) (string, error)
}
