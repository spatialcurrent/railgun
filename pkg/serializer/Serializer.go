package serializer

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/go-simple-serializer/pkg/serializer"
)

type Serializer struct {
	*serializer.Serializer
	alg        string
	dict       []byte
	bufferSize int
	append     bool
	parents    bool
	s3Client   *s3.S3
}

// BufferSize sets the buffer size of the serializer.
func (s *Serializer) BufferSize(bufferSize int) *Serializer {
	s.bufferSize = bufferSize
	return s
}

// Append toggles appending to existing file or overwriting.
func (s *Serializer) Append(append bool) *Serializer {
	s.append = append
	return s
}

// Parents toggles creating parents.
func (s *Serializer) Parents(parents bool) *Serializer {
	s.parents = parents
	return s
}

// S3Client sets the S3 client.
func (s *Serializer) S3Client(s3Client *s3.S3) *Serializer {
	s.s3Client = s3Client
	return s
}

// New returns a new serializer with the given format.
func New(format string, alg string, dict []byte) *Serializer {
	return &Serializer{
		Serializer: serializer.New(format),
		alg:        alg,
		dict:       dict,
		bufferSize: grw.DefaultBufferSize,
		s3Client:   nil,
		append:     false,
		parents:    false,
	}
}

// Deserialize deserializes the input slice of bytes into an object and returns an error, if any.
// Formats jsonl and tags return slices.  If the type is not set, then returns a slice of type []interface{}.
func (s *Serializer) Deserialize(uri string) (interface{}, error) {
	b, err := grw.ReadAllAndClose(&grw.ReadAllAndCloseInput{
		Uri:        uri,
		Alg:        s.alg,
		Dict:       s.dict,
		BufferSize: s.bufferSize,
		S3Client:   s.s3Client,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error reading object at uri %q", uri)
	}
	obj, err := s.Serializer.Deserialize(b)
	if err != nil {
		return nil, errors.Wrapf(err, "error deserializing object at uri %q", uri)
	}
	return obj, nil
}

// Serialize serializes an object into a slice of byte and returns and error, if any.
func (s *Serializer) Serialize(uri string, object interface{}) error {
	b, err := s.Serializer.Serialize(object)
	if err != nil {
		return errors.Wrapf(err, "error serializing object with uri %q", uri)
	}
	err = grw.WriteAllAndClose(&grw.WriteAllAndCloseInput{
		Bytes:    b,
		Uri:      uri,
		Alg:      s.alg,
		Dict:     s.dict,
		Append:   s.append,
		Parents:  s.parents,
		S3Client: s.s3Client,
	})
	if err != nil {
		return errors.Wrapf(err, "error writing object to uri %q", uri)
	}
	return nil
}
