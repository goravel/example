package commands

import (
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type TestConsoleSingleCommand struct{}

func (r *TestConsoleSingleCommand) Signature() string {
	return "test:console-single"
}

func (r *TestConsoleSingleCommand) Description() string {
	return "Test console command with single arguments and all flag types"
}

func (r *TestConsoleSingleCommand) Extend() command.Extend {
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

func (r *TestConsoleSingleCommand) Handle(ctx console.Context) error {
	ctx.Comment("console single command")
	ctx.Info("running single command")
	ctx.Warning("warning")
	ctx.Success("success")

	SetTestConsoleSingleLatest(&TestConsoleSingleCapture{
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
	})

	return nil
}
