package enum

type DeviceType int

const (
	Bluetooth DeviceType = iota
	Wired
	Headphones
)

func (d DeviceType) String() string {
	return [...]string{"Bluetooth", "Wired", "Headphones"}[d]
}

type PlayStrategyType int

const (
	Sequential PlayStrategyType = iota
	Random
	CustomQueue
)
