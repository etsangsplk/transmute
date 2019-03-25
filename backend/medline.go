package backend

import (
	"fmt"
	"github.com/hscells/transmute/fields"
	"github.com/hscells/transmute/ir"
	"github.com/xtgo/set"
	"log"
	"sort"
	"strconv"
	"strings"
)

type MedlineBackend struct {
}

type MedlineQuery struct {
	repr string
}

func (m MedlineQuery) Representation() (interface{}, error) {
	return m.repr, nil
}

func (m MedlineQuery) String() (string, error) {
	return m.repr, nil
}

func (m MedlineQuery) StringPretty() (string, error) {
	return m.repr, nil
}

func compileMedline(q ir.BooleanQuery, level int) (l int, query MedlineQuery) {
	repr := ""
	var op []int
	if q.Keywords == nil && len(q.Operator) == 0 {
		for _, child := range q.Children {
			var comp MedlineQuery
			level, comp = compileMedline(child, level)
			repr += comp.repr
		}
		return level, MedlineQuery{repr: repr}
	}
	for _, child := range q.Children {
		l, comp := compileMedline(child, level)
		repr += comp.repr
		level = l
		op = append(op, l-1)
	}
	for _, keyword := range q.Keywords {
		var mf string
		qs := keyword.QueryString
		if keyword.Exploded {
			qs = "exp " + qs
		}
		if len(keyword.Fields) == 1 && keyword.Fields[0] == fields.MeshHeadings {
			qs += "/"
		} else {
			mapping := map[string][]string{
				"ti,ab,sh": {fields.MeshHeadings, fields.Abstract, fields.Title},
				"ab,sh":    {fields.MeshHeadings, fields.Abstract},
				"ti,sh":    {fields.MeshHeadings, fields.Title},
				"tw":       {fields.Abstract, fields.Title},
				"ab":       {fields.Abstract},
				"ti":       {fields.Title},
				"fs":       {fields.FloatingMeshHeadings},
				"sh":       {fields.MeshHeadings},
				"mh":       {fields.MeSHTerms},
				"pt":       {fields.PublicationType},
				"ed":       {fields.PublicationDate},
				"au":       {fields.Authors},
				"jn":       {fields.Journal},
				"mp":       {fields.AllFields},
				"ti,ab":    {fields.TitleAbstract},
			}
			sort.Strings(keyword.Fields)
			keyword.Fields = set.Strings(keyword.Fields)
			for f, mappingFields := range mapping {
				if len(mappingFields) != len(keyword.Fields) {
					continue
				}
				match := true
				for i, field := range keyword.Fields {
					if field != mappingFields[i] {
						match = false
					}
				}
				if match {
					mf = f
					break
				}
			}
			if len(mf) == 0 {
				log.Println("WARNING: could not map fields: ", keyword)
			}
			qs = fmt.Sprintf("%v.%v.", qs, mf)
		}
		repr += fmt.Sprintf("%v. %v\n", level, qs)
		op = append(op, level)
		level += 1
	}
	if len(op) > 0 {
		// This block of code determines if we can use the short hand version of grouping for medline e.g. or/1-9
		o := op[0]
		asc := true
		for i := 1; i < len(op); i++ {
			if op[i]-1 != o {
				asc = false
				break
			}
			o = op[i]
		}
		if asc && len(op) > 2 {
			repr += fmt.Sprintf("%d. %s/%d-%d\n", level, q.Operator, op[0], op[len(op)-1])
		} else {
			// Otherwise we need to use the long form version.
			ops := make([]string, len(op))
			for i, o := range op {
				ops[i] = strconv.Itoa(o)
			}
			repr += fmt.Sprintf("%v. %v\n", level, strings.Join(ops, fmt.Sprintf(" %v ", q.Operator)))
		}
	}
	level += 1
	return level, MedlineQuery{repr: repr}
}

func (b MedlineBackend) Compile(ir ir.BooleanQuery) (BooleanQuery, error) {
	_, q := compileMedline(ir, 1)
	return q, nil
}

func NewMedlineBackend() MedlineBackend {
	return MedlineBackend{}
}
