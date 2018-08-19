package pgfc

import (
	"github.com/jackc/pgx" // gopkg failed because "internal" lib used
	"gopkg.in/birkirb/loggers.v1"
	"time"
)

func (srv *Server) connectDB(uri string) error {

	srv.log.Debugf("DB connection: %s", uri)
	dbConf, err := pgx.ParseEnvLibpq()
	if err != nil {
		srv.log.Errorf("Unable to parse environment:", err)
		return err
	}
	if uri != "" {
		c, err := pgx.ParseURI("postgres://" + uri)
		if err != nil {
			srv.log.Errorf("Unable to parse connect string:", err)
			return err
		}
		dbConf = dbConf.Merge(c)
	}

	RuntimeParams := make(map[string]string)
	//	RuntimeParams["application_name"] = "dbrpc"
	dbConf.RuntimeParams = RuntimeParams
	if srv.config.Debug {
		dbConf.LogLevel = pgx.LogLevelDebug // LogLevelFromString
	}
	dbConf.Logger = Logger{lg: srv.log, debug: srv.config.Debug}

	var dbh *pgx.ConnPool
	for {
		dbh, err = pgx.NewConnPool(pgx.ConnPoolConfig{
			ConnConfig:     dbConf,
			MaxConnections: 2, //srv.config.Workers,
			AfterConnect: func(conn *pgx.Conn) error {
				if srv.config.Schema != "" {
					srv.log.Debugf("DB searchpath: (%s)", srv.config.Schema)
					_, err = conn.Exec("set search_path = " + srv.config.Schema)
				}
				srv.log.Debugf("Added DB connection")
				// TODO
				//			if cfg.apl.ConfigFunc != "" {
				//				log.Debugf("Loaded DB config from %s", cfg.apl.ConfigFunc)
				//			}
				return err
			},
		})
		if err == nil {
			break
		}
		srv.log.Warnf("DB connect failed: %v", err)
		//
		time.Sleep(time.Second * 5) // sleep & repeat
	}
	if dbh == nil {
		return err
	}
	srv.dbh = dbh
	return nil
}

type Logger struct {
	lg    loggers.Contextual
	debug bool
}

func (l Logger) Log(level pgx.LogLevel, msg string, data map[string]interface{}) {
	if l.debug {
		l.lg.Debugf("DB[%d]:%s (%+v)", level, msg, data)
	}
}
