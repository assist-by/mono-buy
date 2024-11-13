package lib

type SignalType int

const (
	Long SignalType = iota + 1
	Short
	Maybe_Long
	Maybe_Short
	No_Signal
)

var signalTypeNames = [...]string{
	"LONG", "SHORT", "MAYBE LONG", "MAYBE SHORT", "NO SIGNAL",
}

func (t SignalType) String() string {
	idx := t - 1
	if idx < Long || idx > No_Signal {
		return "Unknown"
	}
	return signalTypeNames[idx]
}
