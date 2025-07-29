package paginate

import (
	"github.com/oullin/database/repository/pagination"
	"net/url"
	"strconv"
	"strings"
)

func MakeFrom(url *url.URL, pageSize int) pagination.Paginate {
	page := pagination.MinPage
	values := url.Query()
	path := strings.TrimSpace((*url).Path)

	if values.Get("page") != "" {
		if tPage, err := strconv.Atoi(values.Get("page")); err == nil {
			page = tPage
		}
	}

	if values.Get("limit") != "" {
		if limit, err := strconv.Atoi(values.Get("limit")); err == nil {
			pageSize = limit
		}
	}

	if strings.Contains(path, "categories") && pageSize > pagination.CategoriesMaxLimit {
		pageSize = pagination.CategoriesMaxLimit
	}

	if strings.Contains(path, "posts") && pageSize > pagination.PostsMaxLimit {
		pageSize = pagination.PostsMaxLimit
	}

	if page < pagination.MinPage {
		page = pagination.MinPage
	}

	return pagination.Paginate{
		Page:  page,
		Limit: pageSize,
	}
}
