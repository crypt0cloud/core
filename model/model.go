package model

import (
	"fmt"
)

var conectornamess []string
var conectors map[string]ModelConnector

func RegisterModel(conectorname string, conn ModelConnector){
	if conectors == nil{
		conectors = make(map[string]ModelConnector)
	}
	conectors[conectorname] = conn
}

func Open(conectorname string) (ModelConnector, error){
	val, ok := conectors[conectorname]
	if !ok {
		return nil, fmt.Errorf("Conector doesn't exists")
	}
	return val, nil
}