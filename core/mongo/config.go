package mongo

import (
	"fmt"
	"time"

	"github.com/breezedup/goserver.v2/core"
	"github.com/breezedup/goserver.v2/core/container"
)

var Config = Configuration{
	Dbs: make(map[string]DbConfig),
}

var autoPingInterval time.Duration = 30 * time.Second
var mgoSessions = container.NewSynchronizedMap()
var databases = container.NewSynchronizedMap()
var collections = container.NewSynchronizedMap()

type Configuration struct {
	Dbs map[string]DbConfig
}
type DbConfig struct {
	Host     string
	Database string
	User     string
	Password string
	Safe     mgo.Safe
}

func (c *Configuration) Name() string {
	return "mongo"
}

func (c *Configuration) Init() error {
	//auto ping, ensure net is connected
	go func() {
		for {
			select {
			case <-time.After(autoPingInterval):
				Ping()
			}
		}
	}()
	return nil
}

func (c *Configuration) Close() error {
	sessions := mgoSessions.Items()
	for _, s := range sessions {
		if session, ok := s.(*mgo.Session); ok && session != nil {
			session.Close()
		}
	}
	return nil
}

func init() {
	core.RegistePackage(&Config)
}

func newDBSession(dbc *DbConfig) (s *mgo.Session, err error) {
	login := ""
	if dbc.User != "" {
		login = dbc.User + ":" + dbc.Password + "@"
	}
	host := "localhost"
	if dbc.Host != "" {
		host = dbc.Host
	}

	// http://goneat.org/pkg/labix.org/v2/mgo/#Session.Mongo
	// [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
	url := fmt.Sprintf("mongodb://%s%s/admin", login, host)
	fmt.Println(url)
	session, err := mgo.Dial(url)
	if err != nil {
		return
	}
	session.SetSafe(&dbc.Safe)
	s = session
	return
}

func Ping() {
	var err error
	sessions := mgoSessions.Items()
	for _, s := range sessions {
		if session, ok := s.(*mgo.Session); ok && session != nil {
			err = session.Ping()
			if err != nil {
				session.Refresh()
			}
		}
	}
}

func SetAutoPing(interv time.Duration) {
	autoPingInterval = interv
}

func Database(dbName string) *mgo.Database {
	var dbc DbConfig
	var exist bool
	if dbc, exist = Config.Dbs[dbName]; !exist {
		return nil
	}
	d := databases.Get(dbName)
	if d == nil {
		s, err := newDBSession(&dbc)
		if err != nil {
			fmt.Println("Database:", dbName, " error:", err)
			return nil
		}
		mgoSessions.Set(dbName, s)
		db := s.DB(dbc.Database)
		if db == nil {
			return nil
		}
		databases.Set(dbName, db)
	} else {
		if db, ok := d.(*mgo.Database); ok {
			return db
		}
	}
	return nil
}

func DatabaseC(dbName, collectionName string) *mgo.Collection {
	var dbc DbConfig
	var exist bool
	if dbc, exist = Config.Dbs[dbName]; !exist {
		return nil
	}
	var collMap *container.SynchronizedMap
	cm := collections.Get(dbName)
	if cm == nil {
		collMap = container.NewSynchronizedMap()
		collections.Set(dbName, collMap)
	} else {
		collMap = cm.(*container.SynchronizedMap)
	}
	if collMap == nil {
		return nil
	}

	cc := collMap.Get(collectionName)
	if cc == nil {
		s, err := newDBSession(&dbc)
		if err != nil {
			fmt.Println("DatabaseC:", dbName, collectionName, " error:", err)
			return nil
		}
		key := fmt.Sprintf("%v_%v", dbName, collectionName)
		mgoSessions.Set(key, s)
		db := s.DB(dbc.Database)
		if db == nil {
			return nil
		}
		c := db.C(collectionName)
		if c == nil {
			return nil
		}
		collMap.Set(collectionName, c)
		return c
	} else {
		if c, ok := cc.(*mgo.Collection); ok {
			return c
		}
	}
	return nil
}
