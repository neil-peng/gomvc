package conf

const (
	TABLE_EXAMPLE = "table_example"
)

type DB_CLUSTER struct {
	Db_cluster_tag string
	Db_name        string
	Username       string
	Password       string
	Server         []string
	NameService    string
}

type TABLE_VIEW struct {
	Table_name     string
	Db_cluster_tag string
}

type Conf_Db struct {
	Db_cluseter_switch bool
	Read_timeout_ms    int
	Write_timeout_ms   int
	Timeout_ms         int
	Max_open_conns     int
	Max_idle_conns     int
	Max_conn_timeout   int
	Db_cluster         []*DB_CLUSTER
	Table_view         []*TABLE_VIEW
}

var tableDbTagMap map[string]string

func TableViewToDbCluster(tableView string) string {
	if dbTag, ok := tableDbTagMap[tableView]; ok {
		return dbTag
	}
	return ""
}
