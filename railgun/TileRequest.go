package railgun

import (
	"fmt"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"net/http"
)

type TileRequest struct {
	Header     http.Header
	Collection string
	Z          int
	X          int
	Y          int
	Source     string
	Expression string
	Features   int
}

func (tr TileRequest) String() string {
	str := "requested tile " + fmt.Sprint(tr.Z) + "/" + fmt.Sprint(tr.X) + "/" + fmt.Sprint(tr.Y) + " for collection " + tr.Collection + " from " + tr.Source
	if len(tr.Expression) > 0 {
		str += " with filter " + tr.Expression
	}
	str += " has " + fmt.Sprint(tr.Features) + " features"
	return str
}

func (tr TileRequest) Map() map[string]interface{} {
	return map[string]interface{}{
		"collection": map[string]interface{}{
			"name": tr.Collection,
		},
		"tile": map[string]interface{}{
			"z": tr.Z,
			"x": tr.X,
			"y": tr.Y,
		},
		"source": map[string]interface{}{
			"uri": tr.Source,
		},
		"http": map[string]interface{}{
			"header": tr.Header,
		},
		"dfl": map[string]interface{}{
			"expression": tr.Expression,
		},
		"results": map[string]interface{}{
			"features": tr.Features,
		},
	}
}

func (tr TileRequest) Serialize(format string) (string, error) {
	return gss.SerializeString(tr.Map(), format, []string{}, -1)
}
