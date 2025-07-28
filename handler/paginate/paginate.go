package paginate

import (
	"github.com/oullin/database/repository/pagination"
	"net/url"
	"strconv"
)

func MakeFrom(url url.Values) pagination.Paginate {
	page := pagination.MinPage
	pageSize := pagination.MaxLimit

	if url.Get("page") != "" {
		if tPage, err := strconv.Atoi(url.Get("page")); err == nil {
			page = tPage
		}
	}

	if url.Get("limit") != "" {
		if limit, err := strconv.Atoi(url.Get("limit")); err == nil {
			pageSize = limit
		}
	}

	if page < pagination.MinPage {
		page = pagination.MinPage
	}

	if pageSize > pagination.MaxLimit || pageSize < 1 {
		pageSize = pagination.MaxLimit
	}

	return pagination.Paginate{
		Page:  page,
		Limit: pageSize,
	}
}
