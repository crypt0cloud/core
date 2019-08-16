package query

import (
	"net/http"
	"strconv"
)

func _handleFilters(r *http.Request)(int, int){
	s := r.FormValue("result_size")
	o := r.FormValue("result_offset")

	size, err := strconv.Atoi(s)
	if err != nil {
		return 10, 0
	}

	offset, err := strconv.Atoi(o)
	if err != nil {
		return 10, 0
	}

	return size, offset
}