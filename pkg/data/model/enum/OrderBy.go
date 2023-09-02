package enum

type SortBy int

const (
    SortByCreatedAt = iota
    SortByPrice
)

func (s SortBy) String() string {
    return []string{
        "sort_time",
        "sort_price",
    }[s]
}
