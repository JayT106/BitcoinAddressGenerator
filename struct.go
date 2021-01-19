package main

import "reflect"

type KEYPATH struct {
	ACCOUNT uint32
	CHAIN uint32
	ADDRESS uint32
}

type BIP32PARAM struct {
	SEED string
	PATH KEYPATH
}

// Clear clear the data of a instance especially the importance data like a seed, reduce the possibilities of the malware attack
func Clear(v interface{}) {
	p := reflect.ValueOf(v).Elem()
	p.Set(reflect.Zero(p.Type()))
}