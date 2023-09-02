package enum

type OrderBy int

const (
    OrderByDescending = iota
    OrderByAscending
)

func (s OrderBy) String() string {
    return []string{
        "dsc",
        "asc",
    }[s]
}
