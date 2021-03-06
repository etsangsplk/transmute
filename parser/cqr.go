package parser

import (
	"bytes"
	"encoding/json"
	"github.com/hscells/cqr"
	"github.com/hscells/transmute/fields"
	"github.com/hscells/transmute/ir"
	"log"
)

// CQRTransformer is an implementation of a query transformer for CQR queries.
type CQRTransformer struct{}

// TransformSingle is unused for this parser.
func (c CQRTransformer) TransformSingle(query string, mapping map[string][]string) ir.Keyword {
	return ir.Keyword{}
}

// transformSingle maps CQR keywords to ir keywords.
func transformSingle(rep map[string]interface{}, mapping map[string][]string) ir.Keyword {
	var queryFields []string
	if _, ok := rep["fields"]; ok && rep["fields"] != nil {
		for _, f := range rep["fields"].([]interface{}) {
			if v, ok := mapping[f.(string)]; ok {
				queryFields = append(queryFields, v...)
			} else {
				queryFields = append(queryFields, f.(string))
			}
		}
	} else {
		queryFields = mapping["default"]
	}

	if len(queryFields) == 0 || rep["fields"] == nil {
		queryFields = mapping["default"]
	}

	var exploded, truncated bool
	options := make(map[string]interface{})
	if o, ok := rep["options"].(map[string]interface{}); ok {
		if v, ok := rep["options"].(map[string]interface{})["exploded"]; ok {
			exploded = v.(bool)
		}
		if v, ok := rep["options"].(map[string]interface{})["truncated"]; ok {
			truncated = v.(bool)
		}
		options = o
	}

	query := ""
	if rep["query"] != nil {
		query = rep["query"].(string)
	}

	return ir.Keyword{
		QueryString: query,
		Fields:      queryFields,
		Exploded:    exploded,
		Truncated:   truncated,
		Options:     options,
	}
}

// transformNested transforms the CQR nested queries.
func transformNested(rep map[string]interface{}, mapping map[string][]string) ir.BooleanQuery {
	q := ir.BooleanQuery{Children: []ir.BooleanQuery{}, Keywords: []ir.Keyword{}}

	if rep["options"] != nil {
		q.Options = rep["options"].(map[string]interface{})
	}

	if rep["children"] != nil {
		q.Operator = rep["operator"].(string)
		for _, child := range rep["children"].([]interface{}) {
			cq := child.(map[string]interface{})
			if _, ok := cq["operator"]; !ok {
				q.Keywords = append(q.Keywords, transformSingle(cq, mapping))
			} else {
				q.Children = append(q.Children, transformNested(cq, mapping))
			}
		}
	} else {
		q = ir.BooleanQuery{Operator: cqr.OR, Keywords: []ir.Keyword{transformSingle(rep, mapping)}}
	}

	return q
}

// TransformNested takes a JSON string a parses a CQR object into the ir.
func (c CQRTransformer) TransformNested(query string, mapping map[string][]string) ir.BooleanQuery {
	var queryRep map[string]interface{}
	err := json.Unmarshal(bytes.NewBufferString(query).Bytes(), &queryRep)
	if err != nil {
		log.Println(err)
		return ir.BooleanQuery{}
	}

	return transformNested(queryRep, mapping)
}

// NewCQRParser creates a new parser for CQR queries. This parser makes a lot of assumptions as it assumes the
// structure of this query is perfect.
func NewCQRParser() QueryParser {
	return QueryParser{Parser: CQRTransformer{}, FieldMapping: map[string][]string{"default": {fields.TitleAbstract}}}
}
