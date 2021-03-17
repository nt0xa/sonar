package database

import (
	"fmt"
	"strings"
)

type Pagination struct {
	Count  uint
	After  int64
	Before int64
}

var defaultPagination = Pagination{Count: 10}

func (p *Pagination) IsZero() bool {
	return p.Count == 0 && p.After == 0 && p.Before == 0
}

func (p *Pagination) IsForward() bool {
	return p.After != 0
}

func (p *Pagination) IsBackward() bool {
	return p.Before != 0
}

func condPrefix(query string) string {
	var prefix string
	if strings.Contains(query, "WHERE") {
		prefix = " AND"
	} else {
		prefix = " WHERE"
	}
	return prefix
}

func nextPage(query string, col string, params map[string]interface{}, after int64) (string, map[string]interface{}) {
	query += condPrefix(query) + fmt.Sprintf(" %s > :after", col)
	params["after"] = after
	return query, params
}

func prevPage(query string, col string, params map[string]interface{}, before int64) (string, map[string]interface{}) {
	query = fmt.Sprintf(
		"SELECT * FROM (%s%s %s < :before ORDER BY id ASC) AS p",
		query,
		condPrefix(query),
		col,
	)
	params["before"] = before
	return query, params
}

func paginate(query string, col string, params map[string]interface{}, p Pagination) (string, map[string]interface{}) {
	if p.IsForward() {
		query, params = nextPage(query, col, params, p.After)
	} else if p.IsBackward() {
		query, params = prevPage(query, col, params, p.Before)
	}

	query += fmt.Sprintf(" ORDER BY %s DESC", col)

	if p.Count != 0 {
		query += " LIMIT :count"
		params["count"] = p.Count
	}

	return query, params
}
