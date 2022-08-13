package translations

// Protocol is used to determine the protocol based on the k8s port name.
type Protocol int

const (
	PROTOCOL_TCP = iota
	PROTOCOL_HTTP
	PROTOCOL_REDIS
)

// KubeMapper wraps all conversion functions and allows configuring default values.
type KubeMapper struct{}
