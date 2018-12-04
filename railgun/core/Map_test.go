// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package core

import (
	//"github.com/pkg/errors"
	"reflect"
	//"strings"
	"testing"
)

func TestMap(t *testing.T) {

	fbar := &User{
		Id:           "1",
		UserName:     "fbar",
		FirstName:    "Foo",
		LastName:     "Bar",
		EmailAddress: "fbar@example.com",
	}

	private := map[string]interface{}{
		"id":        "1",
		"username":  "fbar",
		"firstname": "Foo",
		"lastname":  "Bar",
		"email":     "fbar@example.com",
	}

	public := map[string]interface{}{
		"id":        "1",
		"username":  "fbar",
		"firstname": "Foo",
		"lastname":  "Bar",
	}

	if got := Map(fbar, []string{"public", "private"}); !reflect.DeepEqual(got, private) {
		t.Errorf("Map(%v, %v) == %v, want %v", fbar, []string{"public", "private"}, got, private)
	}

	if got := Map(fbar, []string{"public"}); !reflect.DeepEqual(got, public) {
		t.Errorf("Map(%v, %v) == %v, want %v", fbar, []string{"public"}, got, public)
	}

}
