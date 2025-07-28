package pagination

import "gorm.io/gorm"

func Count[T *int64](numItems T, query *gorm.DB, session *gorm.Session, distinct string) error {
	sql := query.
		Session(session).  // clone the based query.
		Distinct(distinct) // remove duplicated; if any to get the actual count.

	if sql.Count(numItems).Error != nil {
		return sql.Error
	}

	return nil
}
