package backend

import (
	"encoding/json"
	"github.com/hscells/cqr"
	"github.com/hscells/transmute/ir"
)

// CommonQueryRepresentationQuery is the transmute wrapper for CQR.
type CommonQueryRepresentationQuery struct {
	repr cqr.CommonQueryRepresentation
}

// CommonQueryRepresentationBackend is the backend for compiling transmute ir into CQR.
type CommonQueryRepresentationBackend struct{}

// Representation returns the CQR.
func (q CommonQueryRepresentationQuery) Representation() interface{} {
	return q.repr
}

// String returns a JSON-encoded representation of the cqr.
func (q CommonQueryRepresentationQuery) String() string {
	b, _ := json.Marshal(q.repr)
	return string(b)
}

// StringPretty returns a pretty-printed JSON-encoded representation of the cqr.
func (q CommonQueryRepresentationQuery) StringPretty() string {
	b, _ := json.MarshalIndent(q.repr, "", "    ")
	return string(b)
}

// Compile transforms the transmute ir into CQR. The CQR is slightly different to the transmute ir, in that the
// depth of the children is different. Take note of how the children of a transmute ir differs from the children of CQR.
func (b CommonQueryRepresentationBackend) Compile(q ir.BooleanQuery) BooleanQuery {
	children := []cqr.CommonQueryRepresentation{}
	for _, keyword := range q.Keywords {
		k := cqr.NewKeyword(keyword.QueryString, keyword.Fields...).
			SetOption("exploded", keyword.Exploded).
			SetOption("truncated", keyword.Truncated)
		children = append(children, k)
	}
	for _, child := range q.Children {
		subChildren := []cqr.CommonQueryRepresentation{}
		for _, subChild := range child.Children {
			cqrSub := b.Compile(subChild).(CommonQueryRepresentationQuery).repr
			subChildren = append(subChildren, cqrSub)
		}
		for _, keyword := range child.Keywords {
			k := cqr.NewKeyword(keyword.QueryString, keyword.Fields...).
				SetOption("exploded", keyword.Exploded).
				SetOption("truncated", keyword.Truncated)
			subChildren = append(subChildren, k)
		}
		children = append(children, cqr.NewBooleanQuery(child.Operator, subChildren))
	}
	repr := cqr.NewBooleanQuery(q.Operator, children)
	return CommonQueryRepresentationQuery{repr: repr}
}

// NewCQRBackend returns a new CQR backend.
func NewCQRBackend() CommonQueryRepresentationBackend {
	return CommonQueryRepresentationBackend{}
}

func NewCQRQuery(query cqr.CommonQueryRepresentation) CommonQueryRepresentationQuery {
	return CommonQueryRepresentationQuery{repr: query}
}