// package pgfc holds pg functions call methods
package pgfc

import (
	"fmt"
	"strings"
	"sync"

	//	"github.com/pkg/errors"

	"github.com/jackc/pgx"
	"gopkg.in/birkirb/loggers.v1"
)

// Config defines local application flags
type Config struct {
	Schema string `long:"db_schema" default:"" description:"Database functions schema name or comma delimited list"`
	Debug  bool   `long:"db_debug" description:"Debug DB operations"`

	InDefFunc     string `long:"db_indef" default:"func_args" description:"Argument definition function"`
	OutDefFunc    string `long:"db_outdef" default:"func_result" description:"Result row definition function"`
	IndexFunc     string `long:"db_index" default:"index" description:"Available functions list"`
	ArgSyntax     string `long:"db_arg_syntax" default:":=" description:"Default named args syntax (:= or =>)"`
	ArgTrimPrefix string `long:"db_arg_prefix" default:"a_" description:"Trim prefix from arg name"`
}

// InDef holds function argument attributes
type InDef struct {
	Name     string
	Type     string
	Required bool
	// Check    string `json:"check,omitempty" sql:"-"` // validate argument
	Default *string `json:",omitempty"`
	Anno    *string `json:",omitempty"`
}

// OutDef holds function result attributes
type OutDef struct {
	Name string
	Type string
	Anno *string `json:",omitempty"`
}

// Method holds method attributes
type Method struct {
	Name     string
	Class    string
	Func     string
	Anno     string
	IsRO     bool
	IsSet    bool
	IsStruct bool
	Sample   *string           `json:",omitempty"`
	Result   *string           `json:",omitempty"`
	In       *map[string]InDef `json:",omitempty"`
	Out      *[]OutDef         `json:",omitempty"`
}

// DB holds RPC methods
type Server struct {
	dbh       *pgx.ConnPool
	config    Config
	log       loggers.Contextual
	methods   *map[string]Method
	internals map[string]string // aliases of Index,inDef,outDef
	mux       sync.RWMutex
}

const (
	// err = rows.Scan(&r.Name, &r.Class, &r.Func, &r.Anno, &r.Sample, &r.Result, &r.IsRO, &r.IsSet, &r.IsStruct)
	SQLMethod = "select code, nspname, proname, anno, sample, result, is_ro, is_set, is_struct from %s($1)"
	// err = rows.Scan(&r.Name, &r.Type, &r.Required, &r.Default, &r.Anno)
	SQLInArgs = "select arg, type, required, def_val, anno from %s($1)"
	// err = rows.Scan(&r.Name, &r.Type, &r.Anno)
	SQLOutArgs = "select arg, type, anno from %s($1)"
)

func NewServer(cfg Config, log loggers.Contextual, uri string) (*Server, error) {

	srv := Server{log: log, config: cfg}
	err := srv.connectDB(uri)
	if err != nil {
		return nil, err
	}
	err = srv.loadMethods(nil)
	if err != nil {
		return nil, err
	}
	return &srv, nil
}

func (srv *Server) Methods() map[string]Method {
	srv.mux.RLock()
	defer srv.mux.RUnlock()
	return *srv.methods
}

// MethodIsRO returns true if method exists and read-only
func (srv *Server) MethodIsRO(method string) bool {
	srv.mux.RLock()
	methods := *srv.methods
	srv.mux.RUnlock()
	m, ok := methods[method]
	if !ok {
		return false
	}
	return m.IsRO
}

func (srv *Server) loadMethods(nsp *string) error {
	sql := fmt.Sprintf(SQLMethod, srv.config.IndexFunc)
	rows, err := srv.dbh.Query(sql, nsp)
	if err != nil {
		return err
	}
	defer rows.Close()
	rv := map[string]Method{}
	for rows.Next() {
		var r Method
		err = rows.Scan(&r.Name, &r.Class, &r.Func, &r.Anno, &r.Sample, &r.Result, &r.IsRO, &r.IsSet, &r.IsStruct)
		if err != nil {
			return err
		}
		r.In, err = srv.loadInArgs(r.Name)
		if err != nil {
			return err
		}
		if r.IsStruct {
			r.Out, err = srv.loadOutArgs(r.Name)
			if err != nil {
				return err
			}
		}
		rv[r.Name] = r
		/*
			funcName := r.Class + '.' + r.Func
			if funcName == s.config.IndexFunc {
				s.internals[r.Name] = "index"
			} else if funcName == s.config.InDefFunc {
				s.internals[r.Name] = "in"
			} else if funcName == s.config.OutDefFunc {
				s.internals[r.Name] = "out"
			}
		*/
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	srv.mux.Lock()
	srv.methods = &rv
	srv.mux.Unlock()
	return nil
}

func (srv *Server) loadInArgs(method string) (*map[string]InDef, error) {
	sql := fmt.Sprintf(SQLInArgs, srv.config.InDefFunc)
	rows, err := srv.dbh.Query(sql, method)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rv := map[string]InDef{}
	for rows.Next() {
		var r InDef
		err = rows.Scan(&r.Name, &r.Type, &r.Required, &r.Default, &r.Anno)
		if err != nil {
			return nil, err
		}
		rv[strings.TrimPrefix(r.Name, srv.config.ArgTrimPrefix)] = r
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return &rv, nil
}

func (srv *Server) loadOutArgs(method string) (*[]OutDef, error) {
	sql := fmt.Sprintf(SQLOutArgs, srv.config.OutDefFunc)
	rows, err := srv.dbh.Query(sql, method)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rv := []OutDef{}
	for rows.Next() {
		var r OutDef
		err = rows.Scan(&r.Name, &r.Type, &r.Anno)
		if err != nil {
			return nil, err
		}
		rv = append(rv, r)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return &rv, nil
}
