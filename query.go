package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func query_handlers() {
	//pq := new(PublicQuery)
	//http.HandleFunc("/query/v1/blocks",apihandlers.Recover(pq.ListBlocks))
	//http.HandleFunc("/query/v1/block_transactions",apihandlers.Recover(pq.ListBlockTransacions))
	//http.HandleFunc("/query/v1/group_transactions",apihandlers.Recover(pq.ListGroupTransacions))

}

func _handleFilters(r *http.Request) (int, int) {
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

type PublicQuery struct {
}

func (pq PublicQuery) ListBlocks(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")

	size, offset := _handleFilters(r)

	blocks := db.GetBlocksByOffset(size, offset)
	json.NewEncoder(w).Encode(blocks)
}

func (pq PublicQuery) ListBlockTransacions(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")

	blockid := r.FormValue("block")
	if blockid == "" {
		panic(fmt.Errorf("In parameters"))
	}

	getall_str := r.FormValue("getall")
	getall := false
	if getall_str != "" {
		getall = true
	}

	size, offset := _handleFilters(r)

	blocks := db.GetBlockTransactions(blockid, size, offset, getall)
	json.NewEncoder(w).Encode(blocks)
}

func (pq PublicQuery) ListGroupTransacions(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")

	group := r.FormValue("group")
	if group == "" {
		panic(fmt.Errorf("In parameters"))
	}

	getall_str := r.FormValue("getall")
	getall := false
	if getall_str != "" {
		getall = true
	}

	size, offset := _handleFilters(r)

	blocks := db.GetGroupTransactions(group, size, offset, getall)
	json.NewEncoder(w).Encode(blocks)
}
