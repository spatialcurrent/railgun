package railgun

import (
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-reader/reader"
	"strings"
)

func ParseDfl(dflUri string, dflExpression string) (dfl.Node, error) {
	var dflNode dfl.Node

	if len(dflUri) > 0 {
		f, _, err := reader.OpenResource(dflUri, "none", 4096, false, nil)
		if err != nil {
			return nil, errors.Wrap(err, "Error opening dfl file")
		}
		content, err := f.ReadAll()
		if err != nil {
			return nil, errors.Wrap(err, "Error reading all from dfl file")
		}
		dflExpression = strings.TrimSpace(dfl.RemoveComments(string(content)))
	}

	if len(dflExpression) > 0 {
		n, err := dfl.Parse(dflExpression)
		if err != nil {
			return nil, errors.Wrap(err, "Error parsing dfl node.")
		}
		dflNode = n.Compile()
	}

	return dflNode, nil
}
