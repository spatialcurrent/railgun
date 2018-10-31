package railgun

import (
	"github.com/spatialcurrent/go-dfl/dfl"
)

type DataStore struct {
	Uri         dfl.Node
	Format      string
	Compression string
	Extent      []float64
}
