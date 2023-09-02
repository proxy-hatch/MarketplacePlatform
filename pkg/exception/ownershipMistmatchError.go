package exception

import "fmt"

type ListingDoesNotExistException struct {
    Context string
    Err     error
}

func NewListingDoesNotExistException(message string, err error) *ListingDoesNotExistException {
    return &ListingDoesNotExistException{
        Context: message,
        Err:     err,
    }
}

func (e *ListingDoesNotExistException) Error() string {
    return fmt.Sprintf("ListingDoesNotExistException: %s: %v", e.Context, e.Err)
}
