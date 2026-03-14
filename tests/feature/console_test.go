package feature

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"goravel/app/console/commands"
	"goravel/app/facades"
	"goravel/tests"
)

type ConsoleTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestConsoleTestSuite(t *testing.T) {
	suite.Run(t, new(ConsoleTestSuite))
}

func (s *ConsoleTestSuite) SetupTest() {
	commands.ResetConsoleCaptures()
}

func (s *ConsoleTestSuite) TearDownTest() {
	commands.ResetConsoleCaptures()
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

	capture := commands.GetConsoleSingleLatest()
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

	capture := commands.GetConsoleSliceLatest()
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

	capture := commands.GetConsoleSingleLatest()
	s.Require().NotNil(capture)
	s.Equal("call", capture.OptionString)
	s.InDelta(float32(1.25), capture.ArgumentFloat32, 0.00001)
	s.True(capture.ArgumentTimestamp.Equal(timestamp))
}
