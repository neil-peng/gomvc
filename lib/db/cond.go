package db

import (
	"fmt"
)

type Cond struct {
	ands           []string
	ors            []string
	orderFieldAsc  []string
	orderFieldDesc []string
	limits         string
	groupby        []string
	having         string
}

func (c *Cond) and(format string, a ...interface{}) {
	if len(a) == 0 {
		c.ands = append(c.ands, format)
	} else {
		c.ands = append(c.ands, fmt.Sprintf(format, a...))
	}
}

func (c *Cond) or(format string, a ...interface{}) {
	if len(a) == 0 {
		c.ors = append(c.ors, format)
	} else {
		c.ors = append(c.ors, fmt.Sprintf(format, a...))
	}
}

func (c *Cond) limit(start int, offset int) {
	c.limits = fmt.Sprintf(" LIMIT %d, %d", start, offset)
}

func (c *Cond) OrderbyAsc(field ...string) {
	if len(field) == 0 {
		return
	}
	c.orderFieldAsc = append(c.orderFieldAsc, field...)
}

func (c *Cond) OrderbyDesc(field ...string) {
	if len(field) == 0 {
		return
	}
	c.orderFieldDesc = append(c.orderFieldDesc, field...)
}

func (c *Cond) Groupby(field ...string) {
	if len(field) == 0 {
		return
	}
	c.groupby = append(c.groupby, field...)
}

func (c *Cond) Having(format string, a ...interface{}) {
	if len(a) == 0 {
		c.having = fmt.Sprintf(" HAVING %s", format)
	} else {
		c.ands = append(c.ands, fmt.Sprintf(format, a...))
		c.having = fmt.Sprintf(" HAVING %s", fmt.Sprintf(format, a...))
	}
}

/*sql select format
SELECT  [DISTINCT | ALL] {* | select_list}
FROM {table_name [alias] | view_name}
    [{table_name [alias]  | view_name}]...
[WHERE  condition]
[GROUP BY  condition_list]
[HAVING  condition]
[ORDER BY  {column_name | column_#  [ ASC | DESC ] } ...
*/
func (c *Cond) format() string {
	var partSql string
	for _, andItem := range c.ands {
		if len(partSql) == 0 {
			partSql = andItem
		} else {
			partSql += " AND " + andItem
		}
	}

	for _, orItem := range c.ors {
		if len(partSql) == 0 {
			partSql = orItem
		} else {
			partSql += " OR " + orItem
		}
	}

	var groupbySql string
	if len(c.groupby) > 0 {
		for i, groupbyField := range c.groupby {
			if i == 0 {
				groupbySql += " GROUP BY " + groupbyField
			} else {
				groupbySql += " , " + groupbyField
			}
		}
	}

	if len(groupbySql) > 0 {
		partSql += groupbySql
	}

	if len(c.having) > 0 {
		partSql += c.having
	}

	var orderPartSql string
	if len(c.orderFieldAsc) > 0 {
		for i, ascField := range c.orderFieldAsc {
			if i == 0 {
				orderPartSql += " ORDER BY " + ascField
			} else {
				orderPartSql += " , " + ascField
			}
		}
		orderPartSql += " ASC "
	}

	if len(c.orderFieldDesc) > 0 {
		for i, descField := range c.orderFieldDesc {
			if i == 0 {
				if len(orderPartSql) == 0 {
					orderPartSql += " ORDER BY " + descField
				} else {
					orderPartSql += " , " + descField
				}
			} else {
				orderPartSql += " , " + descField
			}
		}
		orderPartSql += " DESC "
	}

	if len(orderPartSql) > 0 {
		partSql += orderPartSql
	}

	if len(c.limits) > 0 {
		partSql += c.limits
	}
	return partSql
}
