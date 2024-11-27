package testmodel

import "github.com/fasibio/safe"

type Model1 struct {
	A              string
	A1             safe.Option[string]
	B              int
	B1             safe.Option[int]
	C              float64
	C1             safe.Option[float64]
	D              Model2
	D1             safe.Option[Model2]
	ArrayOfPointer safe.Option[[]*Model2]
}
