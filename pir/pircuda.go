//+build !nocuda,!travis

package pir

import (
	_ "github.com/privacylab/talek/pir/pircuda"
)
