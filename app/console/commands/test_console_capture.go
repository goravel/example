package commands

import "time"

type TestConsoleSingleCapture struct {
	OptionBool         bool
	OptionFloat64      float64
	OptionFloat64Slice []float64
	OptionInt          int
	OptionIntSlice     []int
	OptionInt64        int64
	OptionInt64Slice   []int64
	OptionString       string
	OptionStringSlice  []string

	OptionDefaultString string
	OptionDefaultInt    int
	OptionDefaultBool   bool

	ArgumentFloat32   float32
	ArgumentFloat64   float64
	ArgumentInt       int
	ArgumentInt8      int8
	ArgumentInt16     int16
	ArgumentInt32     int32
	ArgumentInt64     int64
	ArgumentUint      uint
	ArgumentUint8     uint8
	ArgumentUint16    uint16
	ArgumentUint32    uint32
	ArgumentUint64    uint64
	ArgumentTimestamp time.Time

	ArgumentDefaultString string
	ArgumentDefaultInt    int
}

type TestConsoleSliceCapture struct {
	ArgumentStringSlice    []string
	ArgumentFloat32Slice   []float32
	ArgumentFloat64Slice   []float64
	ArgumentIntSlice       []int
	ArgumentInt8Slice      []int8
	ArgumentInt16Slice     []int16
	ArgumentInt32Slice     []int32
	ArgumentInt64Slice     []int64
	ArgumentUintSlice      []uint
	ArgumentUint8Slice     []uint8
	ArgumentUint16Slice    []uint16
	ArgumentUint32Slice    []uint32
	ArgumentUint64Slice    []uint64
	ArgumentTimestampSlice []time.Time

	MissingStringSliceIsNil  bool
	MissingFloat32SliceIsNil bool
	MissingTimestampIsNil    bool
}

var (
	TestConsoleSingleLatest *TestConsoleSingleCapture
	TestConsoleSliceLatest  *TestConsoleSliceCapture
)

func ResetTestConsoleCaptures() {
	TestConsoleSingleLatest = nil
	TestConsoleSliceLatest = nil
}
