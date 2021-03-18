package database

import (
	"fmt"
)

type Page struct {
	Count  uint
	After  int64
	Before int64
}

var defaultPagination = Page{Count: 10}

func (p *Page) IsZero() bool {
	return p.Count == 0 && p.After == 0 && p.Before == 0
}

func (p *Page) IsForward() bool {
	return p.After != 0
}

func (p *Page) IsBackward() bool {
	return p.Before != 0
}

func paginate(
	query string,
	params map[string]interface{},
	column string,
	prefix string,
	order string,
	p Page,
) (string, map[string]interface{}) {

	var (
		cmp string
		ord string

		page bool
	)

	if p.IsForward() {
		page = true
		cmp, ord = ">", "ASC"
		params["paging"] = p.After
	} else if p.IsBackward() {
		page = true
		cmp, ord = "<", "DESC"
		params["paging"] = p.Before
	}

	if page {
		query = fmt.Sprintf(
			"SELECT * FROM (%[1]s %[2]s %[3]s %[4]s :paging ORDER BY %[3]s %[5]s LIMIT :count) AS p",
			query,  // subquery
			prefix, // WHERE/AND
			column, // id
			cmp,    // </>
			ord,    // ASC/DESC
		)
	}

	params["count"] = p.Count

	query += fmt.Sprintf(" ORDER BY %s %s", column, order)

	if !page {
		query += " LIMIT :count"
	}

	return query, params
}
