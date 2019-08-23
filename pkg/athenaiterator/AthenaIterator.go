// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package athenaiterator

import (
	"io"
)

import (
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/pkg/errors"
)

type AthenaIterator struct {
	Client           *athena.Athena
	QueryExecutionId *string
	ResultSet        *athena.ResultSet
	NextToken        *string
	Cursor           int
	Count            int
	Limit            int
}

func (it *AthenaIterator) Next() ([]byte, error) {

	if it.Limit > 0 && it.Count >= it.Limit {
		return make([]byte, 0), io.EOF
	}

	if it.Cursor >= len(it.ResultSet.Rows) {
		if it.NextToken == nil {
			return make([]byte, 0), io.EOF
		} else {
			getQueryResultsOutput, err := it.Client.GetQueryResults(&athena.GetQueryResultsInput{
				QueryExecutionId: it.QueryExecutionId,
				NextToken:        it.NextToken,
			})
			if err != nil {
				return make([]byte, 0), errors.Wrap(err, "error getting results from athena query")
			}
			it.ResultSet = getQueryResultsOutput.ResultSet
			it.NextToken = getQueryResultsOutput.NextToken
			it.Cursor = 0
		}
	}

	if len(it.ResultSet.Rows) == 0 {
		return make([]byte, 0), io.EOF
	}

	str := *(it.ResultSet.Rows[it.Cursor].Data[0].VarCharValue)
	it.Cursor += 1
	it.Count += 1

	return []byte(str), nil
}

func New(client *athena.Athena, queryExecutionId *string, limit int) (*AthenaIterator, error) {

	it := &AthenaIterator{
		Client:           client,
		QueryExecutionId: queryExecutionId,
		ResultSet:        nil,
		NextToken:        nil,
		Cursor:           1, // to skip header
		Count:            0,
		Limit:            limit,
	}

	getQueryResultsOutput, err := it.Client.GetQueryResults(&athena.GetQueryResultsInput{
		QueryExecutionId: it.QueryExecutionId,
		NextToken:        it.NextToken,
	})
	if err != nil {
		return it, errors.Wrap(err, "error getting results from athena query")
	}
	it.ResultSet = getQueryResultsOutput.ResultSet
	it.NextToken = getQueryResultsOutput.NextToken
	return it, nil
}
