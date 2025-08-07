package errs

import (
	"fmt"
	"net/http"
)

type NotFound struct {
	Name string
}

func (e NotFound) Error() string {
	if e.Name == "" {
		return "not found"
	}

	return fmt.Sprintf("%s not found", e.Name)
}

func (a NotFound) HttpStatusCode() int {
	return http.StatusNotFound
}
