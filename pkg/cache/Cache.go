// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cache

import (
	"fmt"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
)

type Cache struct {
	cache *gocache.Cache
}

func (c *Cache) Get(uri string, format string, compression string, bufferSize int, passphrase string, salt string, s3Client *s3.S3) (bool, interface{}, error) {

	item, found := c.cache.Get(uri)
	if found {
		if t := reflect.TypeOf(item); !(t.Kind() == reflect.Array || t.Kind() == reflect.Slice) {
			return true, item, errors.New("object retrieved from cache was not an array or slice but " + fmt.Sprint(t))
		}
		return true, item, nil
	}

	inputReader, _, err := grw.ReadFromResource(&grw.ReadFromResourceInput{
		Uri:        uri,
		Alg:        compression,
		Dict:       grw.NoDict,
		BufferSize: bufferSize,
		S3Client:   s3Client,
	})
	if err != nil {
		return false, nil, errors.Wrap(err, "error opening resource at uri "+uri)
	}

	inputBytes, err := inputReader.ReadAllAndClose()
	if err != nil {
		return false, nil, errors.New("error reading from resource at uri " + uri)
	}

	/*
		inputBytesEncrypted, err := inputReader.ReadAllAndClose()
		if err != nil {
			return false, nil, errors.New("error reading from resource at uri " + uri)
		}

		inputBytesPlain, err := DecryptInput(inputBytesEncrypted, passphrase, salt)
		if err != nil {
			return false, nil, errors.Wrap(err, "error decoding input")
		}*/

	inputType, err := gss.GetType(inputBytes, format)
	if err != nil {
		return false, nil, errors.Wrap(err, "error getting type for input")
	}

	obj, err := gss.DeserializeBytes(&gss.DeserializeBytesInput{
		Bytes:      inputBytes,
		Format:     format,
		Header:     gss.NoHeader,
		Comment:    gss.NoComment,
		LazyQuotes: false,
		SkipLines:  gss.NoSkip,
		Limit:      gss.NoLimit,
		Type:       inputType,
	})
	if err != nil {
		return false, nil, errors.Wrap(err, "error deserializing input using format "+format)
	}

	c.cache.Set(uri, obj, gocache.DefaultExpiration)

	return false, obj, nil
}

func NewCache() *Cache {
	return &Cache{
		cache: gocache.New(5*time.Minute, 10*time.Minute),
	}
}
