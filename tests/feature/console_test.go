package feature

import (
	"testing"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
)

type ConsoleTestSuite struct {
	suite.Suite

	singleCommand *consoleSingleCommand
	sliceCommand  *consoleSliceCommand
}

type consoleSingleCapture struct {
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

type consoleSingleCommand struct {
	capture *consoleSingleCapture
}

func (r *consoleSingleCommand) Signature() string {
	return "test:console-single"
}

func (r *consoleSingleCommand) Description() string {
	return "Test console command with single arguments and all flag types"
}

func (r *consoleSingleCommand) Extend() command.Extend {
	return command.Extend{
		Flags: []command.Flag{
			&command.BoolFlag{Name: "bool"},
			&command.Float64Flag{Name: "float64"},
			&command.Float64SliceFlag{Name: "float64-slice"},
			&command.IntFlag{Name: "int"},
			&command.IntSliceFlag{Name: "int-slice"},
			&command.Int64Flag{Name: "int64"},
			&command.Int64SliceFlag{Name: "int64-slice"},
			&command.StringFlag{Name: "string"},
			&command.StringSliceFlag{Name: "string-slice"},
		},
		Arguments: []command.Argument{
			&command.ArgumentFloat32{Name: "float32Arg", Required: true},
			&command.ArgumentFloat64{Name: "float64Arg", Required: true},
			&command.ArgumentInt{Name: "intArg", Required: true},
			&command.ArgumentInt8{Name: "int8Arg", Required: true},
			&command.ArgumentInt16{Name: "int16Arg", Required: true},
			&command.ArgumentInt32{Name: "int32Arg", Required: true},
			&command.ArgumentInt64{Name: "int64Arg", Required: true},
			&command.ArgumentUint{Name: "uintArg", Required: true},
			&command.ArgumentUint8{Name: "uint8Arg", Required: true},
			&command.ArgumentUint16{Name: "uint16Arg", Required: true},
			&command.ArgumentUint32{Name: "uint32Arg", Required: true},
			&command.ArgumentUint64{Name: "uint64Arg", Required: true},
			&command.ArgumentTimestamp{Name: "timestampArg", Required: true, Layouts: []string{time.RFC3339}},
		},
	}
}

func (r *consoleSingleCommand) Handle(ctx console.Context) error {
	ctx.Comment("console single command")
	ctx.Info("running single command")
	ctx.Warning("warning")
	ctx.Success("success")

	r.capture = &consoleSingleCapture{
		OptionBool:         ctx.OptionBool("bool"),
		OptionFloat64:      ctx.OptionFloat64("float64"),
		OptionFloat64Slice: ctx.OptionFloat64Slice("float64-slice"),
		OptionInt:          ctx.OptionInt("int"),
		OptionIntSlice:     ctx.OptionIntSlice("int-slice"),
		OptionInt64:        ctx.OptionInt64("int64"),
		OptionInt64Slice:   ctx.OptionInt64Slice("int64-slice"),
		OptionString:       ctx.Option("string"),
		OptionStringSlice:  ctx.OptionSlice("string-slice"),

		OptionDefaultString: ctx.Option("missing-option"),
		OptionDefaultInt:    ctx.OptionInt("missing-option"),
		OptionDefaultBool:   ctx.OptionBool("missing-option"),

		ArgumentFloat32:   ctx.ArgumentFloat32("float32Arg"),
		ArgumentFloat64:   ctx.ArgumentFloat64("float64Arg"),
		ArgumentInt:       ctx.ArgumentInt("intArg"),
		ArgumentInt8:      ctx.ArgumentInt8("int8Arg"),
		ArgumentInt16:     ctx.ArgumentInt16("int16Arg"),
		ArgumentInt32:     ctx.ArgumentInt32("int32Arg"),
		ArgumentInt64:     ctx.ArgumentInt64("int64Arg"),
		ArgumentUint:      ctx.ArgumentUint("uintArg"),
		ArgumentUint8:     ctx.ArgumentUint8("uint8Arg"),
		ArgumentUint16:    ctx.ArgumentUint16("uint16Arg"),
		ArgumentUint32:    ctx.ArgumentUint32("uint32Arg"),
		ArgumentUint64:    ctx.ArgumentUint64("uint64Arg"),
		ArgumentTimestamp: ctx.ArgumentTimestamp("timestampArg"),

		ArgumentDefaultString: ctx.ArgumentString("missing-argument"),
		ArgumentDefaultInt:    ctx.ArgumentInt("missing-argument"),
	}

	return nil
}

type consoleSliceCapture struct {
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

type consoleSliceCommand struct {
	capture *consoleSliceCapture
}

func (r *consoleSliceCommand) Signature() string {
	return "test:console-slice"
}

func (r *consoleSliceCommand) Description() string {
	return "Test console command with slice arguments"
}

func (r *consoleSliceCommand) Extend() command.Extend {
	return command.Extend{
		Arguments: []command.Argument{
			&command.ArgumentStringSlice{Name: "stringSliceArg", Min: 1, Max: 2},
			&command.ArgumentFloat32Slice{Name: "float32SliceArg", Min: 1, Max: 2},
			&command.ArgumentFloat64Slice{Name: "float64SliceArg", Min: 1, Max: 2},
			&command.ArgumentIntSlice{Name: "intSliceArg", Min: 1, Max: 2},
			&command.ArgumentInt8Slice{Name: "int8SliceArg", Min: 1, Max: 2},
			&command.ArgumentInt16Slice{Name: "int16SliceArg", Min: 1, Max: 2},
			&command.ArgumentInt32Slice{Name: "int32SliceArg", Min: 1, Max: 2},
			&command.ArgumentInt64Slice{Name: "int64SliceArg", Min: 1, Max: 2},
			&command.ArgumentUintSlice{Name: "uintSliceArg", Min: 1, Max: 2},
			&command.ArgumentUint8Slice{Name: "uint8SliceArg", Min: 1, Max: 2},
			&command.ArgumentUint16Slice{Name: "uint16SliceArg", Min: 1, Max: 2},
			&command.ArgumentUint32Slice{Name: "uint32SliceArg", Min: 1, Max: 2},
			&command.ArgumentUint64Slice{Name: "uint64SliceArg", Min: 1, Max: 2},
			&command.ArgumentTimestampSlice{Name: "timestampSliceArg", Min: 1, Max: 2, Layouts: []string{time.RFC3339}},
		},
	}
}

func (r *consoleSliceCommand) Handle(ctx console.Context) error {
	r.capture = &consoleSliceCapture{
		ArgumentStringSlice:    ctx.ArgumentStringSlice("stringSliceArg"),
		ArgumentFloat32Slice:   ctx.ArgumentFloat32Slice("float32SliceArg"),
		ArgumentFloat64Slice:   ctx.ArgumentFloat64Slice("float64SliceArg"),
		ArgumentIntSlice:       ctx.ArgumentIntSlice("intSliceArg"),
		ArgumentInt8Slice:      ctx.ArgumentInt8Slice("int8SliceArg"),
		ArgumentInt16Slice:     ctx.ArgumentInt16Slice("int16SliceArg"),
		ArgumentInt32Slice:     ctx.ArgumentInt32Slice("int32SliceArg"),
		ArgumentInt64Slice:     ctx.ArgumentInt64Slice("int64SliceArg"),
		ArgumentUintSlice:      ctx.ArgumentUintSlice("uintSliceArg"),
		ArgumentUint8Slice:     ctx.ArgumentUint8Slice("uint8SliceArg"),
		ArgumentUint16Slice:    ctx.ArgumentUint16Slice("uint16SliceArg"),
		ArgumentUint32Slice:    ctx.ArgumentUint32Slice("uint32SliceArg"),
		ArgumentUint64Slice:    ctx.ArgumentUint64Slice("uint64SliceArg"),
		ArgumentTimestampSlice: ctx.ArgumentTimestampSlice("timestampSliceArg"),

		MissingStringSliceIsNil:  ctx.ArgumentStringSlice("missing-slice") == nil,
		MissingFloat32SliceIsNil: ctx.ArgumentFloat32Slice("missing-slice") == nil,
		MissingTimestampIsNil:    ctx.ArgumentTimestampSlice("missing-slice") == nil,
	}

	return nil
}

func TestConsoleTestSuite(t *testing.T) {
	suite.Run(t, new(ConsoleTestSuite))
}

func (s *ConsoleTestSuite) SetupSuite() {
	s.singleCommand = &consoleSingleCommand{}
	s.sliceCommand = &consoleSliceCommand{}
	facades.Artisan().Register([]console.Command{s.singleCommand, s.sliceCommand})
}

func (s *ConsoleTestSuite) TestRunSingleCommand() {
	timestamp := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	err := facades.Artisan().Run([]string{
		"./main", "artisan", "test:console-single",
		"--bool",
		"--float64=3.14",
		"--float64-slice=1.1", "--float64-slice=2.2",
		"--int=42",
		"--int-slice=5", "--int-slice=6",
		"--int64=64",
		"--int64-slice=65", "--int64-slice=66",
		"--string=goravel",
		"--string-slice=go", "--string-slice=framework",
		"3.14", "6.28", "42", "8", "16", "32", "64", "100", "120", "160", "320", "640", timestamp.Format(time.RFC3339),
	}, false)
	s.NoError(err)

	capture := s.singleCommand.capture
	s.Require().NotNil(capture)

	s.True(capture.OptionBool)
	s.InDelta(3.14, capture.OptionFloat64, 0.00001)
	s.Equal([]float64{1.1, 2.2}, capture.OptionFloat64Slice)
	s.Equal(42, capture.OptionInt)
	s.Equal([]int{5, 6}, capture.OptionIntSlice)
	s.Equal(int64(64), capture.OptionInt64)
	s.Equal([]int64{65, 66}, capture.OptionInt64Slice)
	s.Equal("goravel", capture.OptionString)
	s.Equal([]string{"go", "framework"}, capture.OptionStringSlice)

	s.Equal("", capture.OptionDefaultString)
	s.Equal(0, capture.OptionDefaultInt)
	s.False(capture.OptionDefaultBool)

	s.InDelta(float32(3.14), capture.ArgumentFloat32, 0.00001)
	s.InDelta(6.28, capture.ArgumentFloat64, 0.00001)
	s.Equal(42, capture.ArgumentInt)
	s.Equal(int8(8), capture.ArgumentInt8)
	s.Equal(int16(16), capture.ArgumentInt16)
	s.Equal(int32(32), capture.ArgumentInt32)
	s.Equal(int64(64), capture.ArgumentInt64)
	s.Equal(uint(100), capture.ArgumentUint)
	s.Equal(uint8(120), capture.ArgumentUint8)
	s.Equal(uint16(160), capture.ArgumentUint16)
	s.Equal(uint32(320), capture.ArgumentUint32)
	s.Equal(uint64(640), capture.ArgumentUint64)
	s.True(capture.ArgumentTimestamp.Equal(timestamp))

	s.Equal("", capture.ArgumentDefaultString)
	s.Equal(0, capture.ArgumentDefaultInt)
}

func (s *ConsoleTestSuite) TestRunSliceCommand() {
	timestamp1 := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	timestamp2 := timestamp1.Add(time.Hour)

	err := facades.Artisan().Run([]string{
		"./main", "artisan", "test:console-slice",
		"a", "b",
		"1.1", "2.2",
		"3.3", "4.4",
		"10", "20",
		"11", "22",
		"12", "23",
		"13", "24",
		"14", "25",
		"15", "26",
		"16", "27",
		"17", "28",
		"18", "29",
		"19", "30",
		timestamp1.Format(time.RFC3339), timestamp2.Format(time.RFC3339),
	}, false)
	s.NoError(err)

	capture := s.sliceCommand.capture
	s.Require().NotNil(capture)

	s.Equal([]string{"a", "b"}, capture.ArgumentStringSlice)
	s.Equal([]float32{1.1, 2.2}, capture.ArgumentFloat32Slice)
	s.Equal([]float64{3.3, 4.4}, capture.ArgumentFloat64Slice)
	s.Equal([]int{10, 20}, capture.ArgumentIntSlice)
	s.Equal([]int8{11, 22}, capture.ArgumentInt8Slice)
	s.Equal([]int16{12, 23}, capture.ArgumentInt16Slice)
	s.Equal([]int32{13, 24}, capture.ArgumentInt32Slice)
	s.Equal([]int64{14, 25}, capture.ArgumentInt64Slice)
	s.Equal([]uint{15, 26}, capture.ArgumentUintSlice)
	s.Equal([]uint8{16, 27}, capture.ArgumentUint8Slice)
	s.Equal([]uint16{17, 28}, capture.ArgumentUint16Slice)
	s.Equal([]uint32{18, 29}, capture.ArgumentUint32Slice)
	s.Equal([]uint64{19, 30}, capture.ArgumentUint64Slice)
	s.Len(capture.ArgumentTimestampSlice, 2)
	s.True(capture.ArgumentTimestampSlice[0].Equal(timestamp1))
	s.True(capture.ArgumentTimestampSlice[1].Equal(timestamp2))

	s.True(capture.MissingStringSliceIsNil)
	s.True(capture.MissingFloat32SliceIsNil)
	s.True(capture.MissingTimestampIsNil)
}

func (s *ConsoleTestSuite) TestCallSingleCommand() {
	timestamp := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)

	err := facades.Artisan().Call("test:console-single --string=call 1.25 2.5 3 4 5 6 7 8 9 10 11 12 " + timestamp.Format(time.RFC3339))
	s.NoError(err)

	capture := s.singleCommand.capture
	s.Require().NotNil(capture)
	s.Equal("call", capture.OptionString)
	s.InDelta(float32(1.25), capture.ArgumentFloat32, 0.00001)
	s.True(capture.ArgumentTimestamp.Equal(timestamp))
}
