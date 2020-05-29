package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/neil-peng/gomvc/conf"
	"github.com/neil-peng/gomvc/utils"
	"math/rand"
	"net"
	"time"

	"github.com/go-sql-driver/mysql"
)

type DbViewer interface {
	GetTableView() string
	LogId() string
	BuildFields([]map[string]string) ([]interface{}, error)
}

type Db struct {
	mysqlIns *sql.DB
	cluster  *conf.DB_CLUSTER
}

type QueryFiled []string

type UpdateValues map[string]interface{}

type InsertValues map[string]interface{}

type DbQuery struct {
	db          *Db
	dbv         DbViewer
	result      []map[string]string
	err         error
	field       *Field
	cond        *Cond
	sql         string
	forceMaster bool
}

var ClusterTagToDbMap map[string]*Db

func Init(nameServer utils.NameService) {
	mysql.RegisterDial("nameservice", func(serviceName string) (net.Conn, error) {
		ip, port, err := nameServer.GetServer(serviceName)
		if err != nil {
			utils.Critical("init db fail, name:%s, err:%v", serviceName, err)
			return nil, err
		}
		addr := ip + ":" + port
		nd := net.Dialer{Timeout: time.Duration(conf.Db.Timeout_ms) * time.Millisecond}
		return nd.Dial("tcp", addr)
	})

	ClusterTagToDbMap = make(map[string]*Db)
	for _, cluster := range conf.Db.Db_cluster {
		clusterTag := cluster.Db_cluster_tag
		var mysqlIns *sql.DB
		var err error
		for i := 0; i < conf.RETRY; i++ {
			mysqlIns, err = openMysql(cluster)
			if err == nil {
				break
			}
		}
		if err != nil {
			panic(err)
		}
		ClusterTagToDbMap[clusterTag] = &Db{
			mysqlIns: mysqlIns,
			cluster:  cluster,
		}
	}
	return
}

func openMysql(cluster *conf.DB_CLUSTER) (*sql.DB, error) {
	dbConfig := &mysql.Config{
		User:         cluster.Username,
		Passwd:       cluster.Password,
		DBName:       cluster.Db_name,
		Timeout:      time.Duration(conf.Db.Timeout_ms) * time.Millisecond,
		ReadTimeout:  time.Duration(conf.Db.Read_timeout_ms) * time.Millisecond,
		WriteTimeout: time.Duration(conf.Db.Write_timeout_ms) * time.Millisecond,
	}

	if len(cluster.NameService) != 0 {
		dbConfig.Net = "nameservice"
		dbConfig.Addr = cluster.NameService
	} else {
		dbConfig.Net = "tcp"
		dbConfig.Addr = cluster.Server[rand.Intn(len(cluster.Server))]
	}

	mysqlIns, err := sql.Open("mysql", dbConfig.FormatDSN())
	//mysqlIns.SetConnMaxLifetime(time.Duration(conf.Db.Max_conn_timeout) * time.Millisecond)
	mysqlIns.SetMaxOpenConns(conf.Db.Max_open_conns)
	mysqlIns.SetMaxIdleConns(conf.Db.Max_idle_conns)
	if err != nil {
		utils.Critical("open mysql fail, err:%v", err)
		return nil, err
	}
	utils.Notice("open mysql success, conntimeout:%d, maxopen:%d, maxidle:%d",
		conf.Db.Max_conn_timeout, conf.Db.Max_open_conns, conf.Db.Max_idle_conns)
	return mysqlIns, nil
}

func New(dbv DbViewer) *DbQuery {
	return &DbQuery{
		db:    ClusterTagToDbMap[conf.TableViewToDbCluster(dbv.GetTableView())],
		dbv:   dbv,
		field: &Field{},
		cond:  &Cond{},
	}
}

func (d *DbQuery) Clear() {
	d.result = nil
	d.field = &Field{}
	d.cond = &Cond{}
	d.err = nil
	d.sql = ""
	d.forceMaster = false
}

func (d *DbQuery) ForceQueryMaster(forceMaster bool) *DbQuery {
	d.forceMaster = forceMaster
	return d
}

func (d *DbQuery) Field(f string) *DbQuery {
	d.field.field(f)
	return d
}

func (d *DbQuery) Fields(f ...string) *DbQuery {
	d.field.fields(f)
	return d
}

func (d *DbQuery) Value(v interface{}) *DbQuery {
	d.field.value(v)
	return d
}

func (d *DbQuery) Values(v ...interface{}) *DbQuery {
	d.field.values(v...)
	return d
}

func (d *DbQuery) FieldValue(k string, v interface{}) *DbQuery {
	d.field.fieldValue(k, v)
	return d
}

func (d *DbQuery) FieldRawValue(k string, v interface{}) *DbQuery {
	d.field.fieldValue(k, v)
	d.field.fieldRaw(k)
	return d
}

func (d *DbQuery) FieldValues(k string, v ...interface{}) *DbQuery {
	d.field.fieldValues(k, v...)
	return d
}

func (d *DbQuery) AndCond(format string, a ...interface{}) *DbQuery {
	if len(format) == 0 {
		return d
	}
	d.cond.and(format, a...)
	return d
}

func (d *DbQuery) OrCond(format string, a ...interface{}) *DbQuery {
	if len(format) == 0 {
		return d
	}
	d.cond.or(format, a...)
	return d
}

func (d *DbQuery) SetCond(format string, a ...interface{}) *DbQuery {
	if len(format) == 0 {
		return d
	}
	d.cond.and(format, a...)
	return d
}

func (d *DbQuery) Limit(start int, offset int) *DbQuery {
	if start >= 0 && offset >= 0 {
		d.cond.limit(start, offset)
	}
	return d
}

func (d *DbQuery) OrderbyAsc(fields ...string) *DbQuery {
	d.cond.OrderbyAsc(fields...)
	return d
}

func (d *DbQuery) OrderbyDesc(fields ...string) *DbQuery {
	d.cond.OrderbyDesc(fields...)
	return d
}

func (d *DbQuery) Groupby(fields ...string) *DbQuery {
	d.cond.Groupby(fields...)
	return d
}

func (d *DbQuery) Having(format string, a ...interface{}) *DbQuery {
	d.cond.Having(format, a...)
	return d
}

func (d *DbQuery) WhatToInsertORUpdate() ([]interface{}, error) {
	var fieldValueMap []map[string]string
	fieldValueMap = append(fieldValueMap, d.field.FormatIntoMap())
	utils.Debug("WhatToInsertORUpdate, fieldValueMap:%#v", fieldValueMap)

	return d.dbv.BuildFields(fieldValueMap)
}

func (d *DbQuery) Insert() (affectedNum int64, err error) {
	//var sql string
	if len(d.field.formatFields()) == 0 {
		d.sql = fmt.Sprintf("INSERT INTO %s VALUES (%s)", d.dbv.GetTableView(), d.field.formatValues())
	} else {
		d.sql = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", d.dbv.GetTableView(),
			d.field.formatFields(), d.field.formatValues())
	}
	affectedNum, d.err = d.rawExeccSql()
	if d.err != nil {
		if me, ok := d.err.(*mysql.MySQLError); ok {
			//Duplicate entry for key
			if me.Number == 1062 {
				return 0, errors.New(conf.ERROR_DB_QUERY_DUPLICATE)
			}
		}
		return 0, errors.New(conf.ERROR_DB_QUERY_ERROR)
	}
	return affectedNum, nil
}

func (d *DbQuery) Delete() (affectedNum int, err error) {
	d.sql = fmt.Sprintf("DELETE FROM %s WHERE %s", d.dbv.GetTableView(), d.cond.format())
	d.result, d.err = d.rawQuerySql()
	if d.err != nil {
		return 0, errors.New(conf.ERROR_DB_QUERY_ERROR)
	}
	//todo parse affectedNum
	return 0, nil
}

func (d *DbQuery) Update() (affectedNum int, err error) {
	d.sql = fmt.Sprintf("UPDATE %s SET %s WHERE %s", d.dbv.GetTableView(),
		d.field.formatFieldValues(), d.cond.format())

	d.result, d.err = d.rawQuerySql()
	if d.err != nil {
		return 0, errors.New(conf.ERROR_DB_QUERY_ERROR)
	}
	//todo parse affectedNum
	return 0, nil
}

func (d *DbQuery) Select() ([]map[string]string, error) {
	d.sql = fmt.Sprintf("SELECT %s FROM %s WHERE %s", d.field.formatFields(),
		d.dbv.GetTableView(), d.cond.format())
	d.result, d.err = d.rawQuerySql()
	if d.err != nil {
		return nil, errors.New(conf.ERROR_DB_QUERY_ERROR)
	}
	return d.result, nil
}

func (d *DbQuery) SelectToBuild() ([]interface{}, error) {
	d.sql = fmt.Sprintf("SELECT %s FROM %s WHERE %s", d.field.formatFields(),
		d.dbv.GetTableView(), d.cond.format())
	d.result, d.err = d.rawQuerySql()
	return d.BuildFields()
}

func (d *DbQuery) SelectToCount() (int, error) {
	d.sql = fmt.Sprintf("SELECT %s FROM %s WHERE %s", d.field.formatFields(),
		d.dbv.GetTableView(), d.cond.format())
	d.result, d.err = d.rawQuerySql()
	if d.err != nil {
		return 0, d.err
	}
	return len(d.result), nil
}

func (d *DbQuery) BuildFields() ([]interface{}, error) {
	if d.err != nil {
		return nil, d.err
	}
	return d.dbv.BuildFields(d.result)
}

func (d *DbQuery) Sql() string {
	return d.sql
}

func (d *DbQuery) RawQuerySql(sqlStr string) ([]map[string]string, error) {
	d.sql = sqlStr
	return d.rawQuerySql()
}

func (d *DbQuery) RawExeccSql(sqlStr string) (int64, error) {
	d.sql = sqlStr
	return d.rawExeccSql()
}

func (d *DbQuery) addHint() string {
	commentParam := make(map[string]interface{})
	commentParam["comment"] = 1
	commentParam["log_id"] = d.dbv.LogId()
	if d.forceMaster {
		commentParam["is_master"] = 2
	}
	commentStr, _ := json.Marshal(commentParam)
	return "/*" + string(commentStr) + "*/"
}

func (d *DbQuery) rawExeccSql() (int64, error) {
	cost := time.Now()
	sqlFormat := d.addHint() + d.Sql()
	mysqlIns := d.db.mysqlIns
	logId := d.dbv.LogId()
	var affectedNum int64
	defer func() {
		utils.Info("logid:%v, status:%+v, sql:%s, affectedNum:%d, err:%v, cost:%dus",
			logId, mysqlIns.Stats(), sqlFormat, affectedNum, d.err, time.Since(cost)/time.Microsecond)
	}()

	var result sql.Result
	result, d.err = mysqlIns.Exec(sqlFormat)
	if d.err != nil {
		utils.Warn("logid:%v, exec fail, sql:%s, err:%v", logId, sqlFormat, d.err)
		return 0, d.err
	}
	if affectedNum, d.err = result.RowsAffected(); d.err != nil {
		utils.Warn("logid:%v, exec error] [sql:%s] [err:%v]", logId, sqlFormat, d.err)
		return 0, d.err
	}

	return affectedNum, d.err
}

//查询失败error!=nil; 查询为空map=nil，error=nil
func (d *DbQuery) rawQuerySql() ([]map[string]string, error) {
	cost := time.Now()
	sqlFormat := d.addHint() + d.Sql()
	mysqlIns := d.db.mysqlIns
	logId := d.dbv.LogId()
	defer func() {
		utils.Info("logid:%v, status:%+v, sql:%s, len_res:%d, cost:%dus, err:%v]",
			logId, mysqlIns.Stats(), sqlFormat, len(d.result), time.Since(cost)/time.Microsecond, d.err)
	}()

	var rows *sql.Rows
	var cols []string

	rows, d.err = mysqlIns.Query(sqlFormat)
	if d.err != nil {
		utils.Warn("[logid:%v] [query error] [sql:%s] [err:%v]", logId, sqlFormat, d.err)
		return nil, d.err
	}
	defer rows.Close()
	cols, d.err = rows.Columns()
	if d.err != nil {
		return nil, d.err
	}

	d.result = nil
	value := make([][]byte, len(cols))
	scanArgs := make([]interface{}, len(cols))
	for i := range value {
		scanArgs[i] = &value[i]
	}

	for rows.Next() {
		if d.err = rows.Scan(scanArgs...); d.err != nil {
			d.result = nil
			return d.result, d.err
		}
		row := make(map[string]string)
		for i, v := range value {
			if v != nil {
				row[cols[i]] = string(v)
			}
		}
		d.result = append(d.result, row)
	}
	return d.result, d.err
}
