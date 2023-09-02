package exception

import "fmt"

type OwnershipMismatchException struct {
    Context string
    Err     error
}

func NewOwnershipMismatchException(message string, err error) *OwnershipMismatchException {
    return &OwnershipMismatchException{
        Context: message,
        Err:     err,
    }
}

func (e *OwnershipMismatchException) Error() string {
    return fmt.Sprintf("OwnershipMismatchException: %s: %v", e.Context, e.Err)
}
