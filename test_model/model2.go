package testmodel

import "github.com/fasibio/safe"

type Model2 struct {
	A []string
	B safe.Option[[]string]
	C safe.Option[string]
	D safe.Option[[10]string]
}
