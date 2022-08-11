// Package pagination defines helper models for pagination related functionalities
package pagination

const (
	// PageDefaultNumber int value 1
	PageDefaultNumber int = 1
	// PageDefaultSize int value 10
	PageDefaultSize int = 10
	// PageDefaultSortBy default sortBy string value
	PageDefaultSortBy string = "created_at"
	// PageDefaultSortDirectionDesc default sort direction descending order status
	PageDefaultSortDirectionDesc bool = true
	// PageSortDirectionAscending string value asc
	PageSortDirectionAscending string = "asc"
	// PageSortDirectionDescending string value desc
	PageSortDirectionDescending string = "desc"
)

// Page object for pagination request
type Page struct {
	Number            *int    `json:"number" validate:"required"`
	Size              *int    `json:"size" validate:"required"`
	SortBy            *string `json:"sort_by" validate:"required"`
	SortDirectionDesc *bool   `json:"sort_direction_desc" validate:"required"`
}

// PageInfo holds pagination response info
type PageInfo struct {
	Page            int   `json:"page"`
	Size            int   `json:"size"`
	HasNextPage     bool  `json:"has_next_page"`
	HasPreviousPage bool  `json:"has_previous_page"`
	TotalCount      int64 `json:"total_count"`
}

// NewPage creates a new pagination Page object
func NewPage(n int, s int, sBy string, sDirectionD bool) Page {
	return Page{
		Number:            &n,
		Size:              &s,
		SortBy:            &sBy,
		SortDirectionDesc: &sDirectionD,
	}
}

// NewPageWithDefaultSorting creates a new pagination Page object with system default values
func NewPageWithDefaultSorting(n int, s int) Page {
	return Page{
		Number: &n,
		Size:   &s,
	}
}
