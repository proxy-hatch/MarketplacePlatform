package data

type DataAccess interface {
    Save(item interface{}) error
    Retrieve(id string) (interface{}, error)
    Update(id string, item interface{}) error
    Delete(id string) error
}
