package commands

import (
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type ConsoleSliceCommand struct{}

func (r *ConsoleSliceCommand) Signature() string {
	return "test:console-slice"
}

func (r *ConsoleSliceCommand) Description() string {
	return "Test console command with slice arguments"
}

func (r *ConsoleSliceCommand) Extend() command.Extend {
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

func (r *ConsoleSliceCommand) Handle(ctx console.Context) error {
	SetConsoleSliceLatest(&ConsoleSliceCapture{
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
	})

	return nil
}
