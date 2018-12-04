// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package core

import (
	"context"
	"github.com/spatialcurrent/go-dfl/dfl"
	"reflect"
)

type User struct {
	Id           string `rest:"id, the id of the user" required:"yes" visibility:"{public,private}"`
	UserName     string `rest:"username, the user name of the user" required:"yes" visibility:"{public,private}"`
	FirstName    string `rest:"firstname, the first name of the user" required:"yes" visibility:"{public,private}"`
	LastName     string `rest:"lastname, the last name of the user" required:"yes" visibility:"{public,private}"`
	EmailAddress string `rest:"email, the email address of the user" required:"yes" visibility:"{private}"`
}

var UserType = reflect.TypeOf(User{})
var UserVisibilities = CompileVisibilities(reflect.TypeOf(User{}))

func (u User) Name() string {
	return u.UserName
}

func (u User) Map(ctx context.Context) map[string]interface{} {

	//authorizations:= []string{"public"}
	//if request.Id ==
	//Map(object interface{}, authorizations []string)

	m := map[string]interface{}{}
	return m
}

func (u User) Dfl(ctx context.Context) string {
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range u.Map(ctx) {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	return dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
}
