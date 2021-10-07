package cliu

import "github.com/urfave/cli/v2"

type BoolFlag struct {
	*cli.BoolFlag
	CurrentValue bool
}

func NewBoolFlag(flag *cli.BoolFlag) (ret *BoolFlag) {
	ret = &BoolFlag{BoolFlag: flag}
	flag.Destination = &ret.CurrentValue
	return
}

type StringFlag struct {
	*cli.StringFlag
	CurrentValue string
}

func NewStringFlag(flag *cli.StringFlag) (ret *StringFlag) {
	ret = &StringFlag{StringFlag: flag}
	flag.Destination = &ret.CurrentValue
	return
}

type IntFlag struct {
	*cli.IntFlag
	CurrentValue int
}

func NewIntFlag(flag *cli.IntFlag) (ret *IntFlag) {
	ret = &IntFlag{IntFlag: flag}
	flag.Destination = &ret.CurrentValue
	return
}

type UintFlag struct {
	*cli.UintFlag
	CurrentValue uint
}

func NewUintFlag(flag *cli.UintFlag) (ret *UintFlag) {
	ret = &UintFlag{UintFlag: flag}
	flag.Destination = &ret.CurrentValue
	return
}
