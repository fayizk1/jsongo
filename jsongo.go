// Copyright 2014 Benny Scetbun. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package Jsongo is a simple library to help you build Json without static struct
//
// Source code and project home:
// https://github.com/benny-deluxe/jsongo
//

package jsongo

import (
	"encoding/json"
	"errors"
)

//ErrorKeyAlreadyExist error if a key already exist in current JSONNode
var ErrorKeyAlreadyExist = errors.New("jsongo key already exist")

//ErrorMultipleType error if a JSONNode already got a different type of value
var ErrorMultipleType = errors.New("jsongo this node is already set to a different jsonNodeType")

//ErrorArrayNegativeValue error if you ask for a negative index in an array
var ErrorArrayNegativeValue = errors.New("jsongo negative index for array")

//ErrorArrayNegativeValue error if you ask for a negative index in an array
var ErrorAtUnsupportedType = errors.New("jsongo Unsupported Type as At argument")

//ErrorRetrieveUserValue error if you ask the value of a node that is not a value node
var ErrorRetrieveUserValue = errors.New("jsongo Cannot retrieve node's value which is not of type value")

//ErrorTypeUnmarshaling error if you try to unmarshal something in the wrong type 
var ErrorTypeUnmarshaling = errors.New("jsongo Wrong type when Unmarshaling")

//JSONNode Datastructure to build and maintain Nodes
type JSONNode struct {
	m map[string]*JSONNode
	a []JSONNode
	v interface{}
	t jsonNodeType //Type of that JSONNode 0: Not defined, 1: map, 2: array, 3: value
	dontGenerate bool //dont generate while Unmarshal
}

type jsonNodeType int
const (
	//TypeUndefined is set by default for empty JSONNode
	TypeUndefined jsonNodeType = iota 
	//TypeMap is set when a JSONNode is a Map
	TypeMap
	//TypeArray is set when a JSONNode is an Array
	TypeArray
	//TypeValue is set when a JSONNode is a Value Node
	TypeValue 
)

//At At helps you move through your node by building them on the fly
//val can be string or int values
//string are keys for map in json
//int are index in array in json
func (that *JSONNode) At(val ...interface{}) *JSONNode {
	if len(val) == 0 {
		return that
	}
	switch vv := val[0].(type) {
	case string:
		return that.atMap(vv, val[1:]...)
	case int:
		return that.atArray(vv, val[1:]...)
	}
	panic(ErrorAtUnsupportedType)
}

//atMap return the JSONNode in current map
func (that *JSONNode) atMap(key string, val ...interface{}) *JSONNode {
	if that.t != TypeUndefined && that.t != TypeMap {
		panic(ErrorMultipleType)
	}
	if that.m == nil {
		that.m = make(map[string]*JSONNode)
		that.t = TypeMap
	}
	if next, ok := that.m[key]; ok {
		return next.At(val...)
	}
	that.m[key] = new(JSONNode)
	return that.m[key].At(val...)
}

//atArray return the JSONNode in current TypeArray (and make it grow if necessary)
func (that *JSONNode) atArray(key int, val ...interface{}) *JSONNode {
	if that.t == TypeUndefined {
		that.t = TypeArray
	} else if that.t != TypeArray {
		panic(ErrorMultipleType)
	}
	if key < 0 {
		panic(ErrorArrayNegativeValue)
	}
	if key >= len(that.a) {
		newa := make([]JSONNode, key+1)
		for i := 0; i < len(that.a); i++ {
			newa[i] = that.a[i]
		}
		that.a = newa
	}
	return that.a[key].At(val...)
}

//Map Turn this JSONNode to a map and Create a new element for key
func (that *JSONNode) Map(key string) *JSONNode {
	if that.t != TypeUndefined && that.t != TypeMap {
		panic(ErrorMultipleType)
	}
	if that.m == nil {
		that.m = make(map[string]*JSONNode)
		that.t = TypeMap
	}
	if _, ok := that.m[key]; ok {
		panic(ErrorKeyAlreadyExist)
	}
	that.m[key] = &JSONNode{}
	return that.m[key]
}

//Array Turn this JSONNode to an array and/or set array size (reducing size will make you loose data)
func (that *JSONNode) Array(size int) *[]JSONNode {
	if that.t == TypeUndefined {
		that.t = TypeArray
	} else if that.t != TypeArray {
		panic(ErrorMultipleType)
	}
	if size < 0 {
		panic(ErrorArrayNegativeValue)
	}
	var min int
	if size < len(that.m) {
		min = size
	} else {
		min = len(that.m)
	}
	newa := make([]JSONNode, size)
	for i := 0; i < min; i++ {
		newa[i] = that.a[i]
	}
	that.a = newa
	return &(that.a)
}

//Val Turn this JSONNode to Value type and set that value
func (that *JSONNode) Val(val interface{}) {
	if that.t == TypeUndefined {
		that.t = TypeValue
	} else if that.t != TypeValue {
		panic(ErrorMultipleType)
	}
	that.v = val
}

//Get Return user value as interface{}
func (that *JSONNode) Get() interface{} {
	if that.t != TypeValue {
		panic(ErrorRetrieveUserValue)
	}
	return that.v
}

//Unset Will unset everything in the JSONnode. All the children data will be lost
func (that *JSONNode) Unset() {
	*that = JSONNode{}
}

//UnmarshalDontGenerate set or not if Unmarshall will generate anything in that JSONNode and its children
//val: Setting this to true will avoid generation from Unmarshal but will save the value as interface if the current node is of type Value or Undefined
//recurse: Will set all the children of that JSONNode
func (that *JSONNode) UnmarshalDontGenerate(val bool, recurse bool) {
	that.dontGenerate = val
	if recurse {
		switch that.t {
			case TypeMap:
				for k := range that.m {
					that.m[k].UnmarshalDontGenerate(val, recurse)
				}
			case TypeArray:
				for k := range that.a {
					that.a[k].UnmarshalDontGenerate(val, recurse)
				}
		}
	}
}

//MarshalJSON Make JSONNode a Marshaler Interface compatible
func (that *JSONNode) MarshalJSON() ([]byte, error) {
	var ret []byte
	var err error
	switch that.t {
	case TypeMap:
		ret, err = json.Marshal(that.m)
	case TypeArray:
		ret, err = json.Marshal(that.a)
	case TypeValue:
		ret, err = json.Marshal(that.v)
	default:
		ret, err = json.Marshal(nil)
	}
	if err != nil {
		return nil, err
	}
	return ret, err
}

//UnmarshalJSON Make JSONNode a Unmarshaler Interface compatible
func (that *JSONNode) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if !(that.dontGenerate && that.t == TypeUndefined) {
		if data[0] == '{' {
			if that.t != TypeMap && that.t != TypeUndefined {
				return ErrorTypeUnmarshaling
			}
			tmp := make(map[string]json.RawMessage)
			err := json.Unmarshal(data, &tmp)
			if err != nil {
				return err
			}
			for k := range tmp {
				if _, ok := that.m[k]; ok {
					err := json.Unmarshal(tmp[k], that.m[k])
					if err != nil {
						return err
					}
				} else if !that.dontGenerate {
					err := json.Unmarshal(tmp[k], that.Map(k))
					if err != nil {
						return err
					}
				}
			}
			return nil
		}
		if data[0] == '[' {
			if that.t != TypeArray && that.t != TypeUndefined {
				return ErrorTypeUnmarshaling
			}
			var tmp []json.RawMessage
			err := json.Unmarshal(data, &tmp)
			if err != nil {
				return err
			}
			for i := len(tmp) - 1; i >= 0; i-- {
				if !that.dontGenerate || i < len(that.a) {
					err := json.Unmarshal(tmp[i], that.At(i))
					if err != nil {
						return err
					}
				}
			}
			return nil
		}
	}
	var tmp interface{}
	err :=  json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	that.Val(tmp)
	return nil
}
