package pagination

type Paginate struct {
	Page     int
	Limit    int
	NumItems int64
}

func (a *Paginate) SetNumItems(number int64) {
	a.NumItems = number
}

func (a *Paginate) GetNumItemsAsInt() int64 {
	return a.NumItems
}

func (a *Paginate) GetNumItemsAsFloat() float64 {
	return float64(a.NumItems)
}

func (a *Paginate) GetLimit() int {
	return a.Limit
}
