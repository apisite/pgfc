// Copyright 2018 Aleksei Kovrizhkin <lekovr+apisite@gmail.com>. All rights reserved.

package pgfc

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// Call postgresql stored function
func (srv *Server) Call(r *http.Request,
	method string,
	args *map[string]interface{},
) (interface{}, error) {
	// Lookup method.
	methodSpec, ok := srv.Methods()[method]
	if !ok {
		return nil, errors.New("Method not found")
	}

	if typ, ok := srv.internals[method]; ok {
		// TODO:  return data from memory
		srv.log.Debugf("Request for internal method (%s)", typ)
	}
	var inAssigns []string
	var inVars []interface{}

	if methodSpec.In != nil {
		srv.log.Printf("IN args: %+v", *methodSpec.In)
		for k, v := range *methodSpec.In {
			a, ok := (*args)[k]
			if !ok {
				if v.Required {
					return nil, errors.New("Missed required arg:" + k)
				}
				srv.log.Printf("SKIP1: %s", k)
				continue
			}
			if reflect.ValueOf(a).Kind() == reflect.Ptr {
				if reflect.ValueOf(a).IsNil() {
					if v.Required {
						return nil, errors.New("Missed required arg:" + k)
					}
					srv.log.Printf("SKIP2: %s", k)
					continue
				}
				a = reflect.ValueOf(a).Elem().Interface() // dereference ptr
			}
			inAssigns = append(inAssigns, fmt.Sprintf("%s %s $%d", v.Name, srv.config.ArgSyntax, len(inAssigns)+1))
			inVars = append(inVars, a)
			srv.log.Printf("USE: %s (%+v)", k, a)
		}
	}

	var outCols []string
	if methodSpec.Out != nil {
		for _, v := range *methodSpec.Out {
			outCols = append(outCols, v.Name)
		}
	}

	if methodSpec.Out == nil && methodSpec.Result == nil {
		// no data returned
		sql := fmt.Sprintf("SELECT %s.%s(%s)",
			methodSpec.Class,
			methodSpec.Func,
			strings.Join(inAssigns, ", "),
		)
		ct, err := srv.dbh.Exec(sql, inVars...)
		srv.log.Printf("Rows affected: %d", ct.RowsAffected()) // TODO: Header.Add ?
		return nil, err
	}

	from := ""
	if len(outCols) > 0 {
		from = " from "
	}

	rvDB := [][]interface{}{} //reflect.New(methodSpec.outType).Elem().Interface()

	sql := fmt.Sprintf("select %s%s%s.%s(%s)",
		strings.Join(outCols, ", "),
		from,
		methodSpec.Class,
		methodSpec.Func,
		strings.Join(inAssigns, ", "),
	)
	rows, err := srv.dbh.Query(sql, inVars...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		row, err := rows.Values()
		if err != nil {
			return nil, err
		}
		rvDB = append(rvDB, row)

	}
	rv := []interface{}{}
	for _, v := range rvDB {
		if !methodSpec.IsStruct {
			// transform row from array to scalar
			rv = append(rv, v[0])
		} else {
			// transform array into struct or map
			rowMap := map[string]interface{}{}
			for i, n := range *methodSpec.Out {
				if !reflect.ValueOf(v[i]).IsValid() {
					// skip reference to nil
					continue
				}
				rowMap[n.Name] = v[i]
			}
			rv = append(rv, rowMap)
		}
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	if !methodSpec.IsSet {
		if len(rv) != 1 {
			return nil, errors.New("single row must be returned")
		}
		return &rv[0], nil
	}
	return &rv, nil
}
