package railgun

import (
	"github.com/spatialcurrent/go-dfl/dfl"
	"reflect"
)

type DataStore struct {
	Workspace   Workspace `rest:"workspace, the name of the containing workspace" required:"yes"`
	Name        string    `rest:"name, the unique name of the data store" required:"yes"`
	Title       string    `rest:"title, the title of the data store"`
	Description string    `rest:"description, a verbose description of the data store"`
	Uri         dfl.Node  `rest:"uri, a uri to the data (local or AWS s3)" required:"yes"`
	Format      string    `rest:"format, the format of the data (default inferred from uri)"`
	Compression string    `rest:"compression, the compression of the data (default inferred from uri)"`
	Extent      []float64 `rest:"extent, the extent of the data"`
}

func (ds DataStore) Map() map[string]interface{} {
	return map[string]interface{}{
		"workspace":   ds.Workspace.Name,
		"name":        ds.Name,
		"title":       ds.Title,
		"description": ds.Description,
		"uri":         ds.Uri.Dfl(dfl.DefaultQuotes, false, 0),
		"format":      ds.Format,
		"compression": ds.Compression,
		"extent":      dfl.Literal{Value: ds.Extent}.Dfl(dfl.DefaultQuotes, false, 0),
	}
}

func (ds DataStore) Dfl() string {
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range ds.Map() {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	return dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
}

var DataStoreType = reflect.TypeOf(DataStore{})
