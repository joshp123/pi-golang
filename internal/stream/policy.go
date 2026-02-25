package stream

type Mode string

const (
	ModeDrop  Mode = "drop"
	ModeBlock Mode = "block"
	ModeRing  Mode = "ring"
)

type Policy struct {
	Buffer        int
	Mode          Mode
	EmitDropEvent bool
}
