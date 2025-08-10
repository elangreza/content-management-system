package params

import (
	"errors"
	"fmt"
	"strings"
)

type PaginationParams struct {
	Sorts []string
	Limit int
	Page  int

	// to validate sort keys
	validSortKeys map[string]bool

	// local var. Used for sorting in the DB
	// sortMap        map[string]string
	sortDirections []string
	orderClause    string
}

func (pqr *PaginationParams) Validate() error {
	if pqr.Page < 1 {
		pqr.Page = 1
	}

	if pqr.Limit < 1 {
		pqr.Limit = 10
	}

	if len(pqr.Sorts) > 0 {
		newSorts := []string{}
		for _, sort := range pqr.Sorts {
			if strings.Contains(sort, ",") {
				newSorts = append(newSorts, strings.Split(sort, ",")...)
			} else {
				newSorts = append(newSorts, sort)
			}
		}

		// pqr.sortMap = make(map[string]string)
		pqr.sortDirections = make([]string, len(newSorts))
		for index, sortRaw := range newSorts {
			parts := strings.Split(sortRaw, ":")
			if len(parts) != 2 {
				return fmt.Errorf("%s is not valid sort format", sortRaw)
			}

			value := strings.ToLower(strings.TrimSpace(parts[0]))
			direction := strings.ToLower(strings.TrimSpace(parts[1]))

			if direction != "asc" && direction != "desc" {
				return errors.New("not valid sort direction")
			}

			if _, ok := pqr.validSortKeys[value]; !ok {
				return fmt.Errorf("%s is not valid sort key", value)
			}

			pqr.sortDirections[index] = fmt.Sprintf("%s %s", value, direction)

		}

		pqr.orderClause = strings.Join(pqr.sortDirections, ", ")
	}

	return nil
}

func (pqr *PaginationParams) GetOrderClause() string {
	return pqr.orderClause
}

func (pqr *PaginationParams) setValidSortKey(sortKeys ...string) {
	if pqr.validSortKeys == nil {
		pqr.validSortKeys = make(map[string]bool)
	}

	for _, sortKey := range sortKeys {
		pqr.validSortKeys[sortKey] = true
	}
}
