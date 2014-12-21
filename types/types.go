package types

import (
	"github.com/fatih/structs"
	"reflect"
)

// Values that are exposed to otto must be of a certain kind. They should only consist of basic types such as
// primitives, arrays/slices, and maps. This goes for values that are returned by functions as well. This function
// takes a given go value and turns it into an Atë compatible value.
//
// Generally speaking, numbers, strings and booleans are passed right in. Structs (and pointers to structs)
// are transformed into map[string]interface{} objects. This is done recursively. Arrays and slices are converted
// into arrays of interfaces. If the elements are structs, then they are converted to maps, and a new array is
// created.
//
// This is not the most efficient way of doing things. It will be optimized later, probably as part of a significant
// overhaul of the entire javascript backend.
func ToJsValue(input interface{}) interface{} {
	rv := reflect.ValueOf(input)
	kind := rv.Kind()
	if isPrim(rv) {
		return input
	} else if isPrimPtr(rv) {
		return rv.Elem()
	} else if structs.IsStruct(input) {
		// This handles both structs and pointers to structs.
		return structs.Map(input)
	} else if kind == reflect.Map {
		keys := rv.MapKeys()
		if keys == nil || len(keys) == 0 {
			return make(map[string]interface{}) 
		}
		if keys[0].Kind() != reflect.String {
			panic("Keys in maps that are exported to the javascript runtime are only allowed to be strings.")
		}
		mp := make(map[string]interface{})
		for _, key := range keys {
			// Call this recursively.
			mp[key.String()] = ToJsValue(rv.MapIndex(key))
		}
	} else if kind == reflect.Slice || kind == reflect.Array {
		mp := make([]interface{},rv.Len())
		for i := 0; i < rv.Len(); i++ {
			// Call this recursively.
			mp[i] = ToJsValue(rv.Index(i))
		}
	} else if kind == reflect.Uintptr || kind == reflect.UnsafePointer {
		panic("uintptrs and unsafe pointers can not be exposed to the javascript runtime.")	
	} else if kind == reflect.Complex64 {
		cplx, _ := input.(complex64)
		// Just make maps out of these.
		ret64 := make(map[string]interface{})
		ret64["Real"] = real(cplx)
		ret64["Imag"] = imag(cplx)
		return ret64
	} else if kind == reflect.Complex128 {
		cplx, _ := input.(complex128)
		// Just make maps out of these.
		ret128 := make(map[string]interface{})
		ret128["Real"] = real(cplx)
		ret128["Imag"] = imag(cplx)
		return ret128
	} else if kind == reflect.Ptr {
		// Call this recursively.
		return ToJsValue(rv.Elem())
	} else {
		panic("Unsupported type: " + rv.Kind().String())
	}
	return nil
}

func isPrim(v reflect.Value) bool {
	kind := v.Kind()
	return (kind > reflect.Invalid && kind < reflect.Complex64 && !(kind == reflect.Uintptr)) || kind == reflect.String
}

func isPrimPtr(v reflect.Value) bool {
	if v.Kind() == reflect.Ptr {
		return isPrim(v.Elem())
	}
	return false
}
