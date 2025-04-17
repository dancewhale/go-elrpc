package elrpc

import (
	"fmt"
	"reflect"

	"github.com/dancewhale/go-elrpc/parser"
)

func Decode(sexp string) ([]interface{}, error) {
	sexps, err := DecodeToSExp(sexp)
	if err != nil {
		return nil, err
	}
	ret := make([]interface{}, len(sexps))
	for i, sexp := range sexps {
		ret[i] = sexp.ToValue()
	}
	return ret, nil
}

func Decode1(sexp string) (interface{}, error) {
	sexps, err := DecodeToSExp(sexp)
	if err != nil {
		return nil, err
	}
	if len(sexps) == 0 {
		return nil, nil
	}
	return sexps[0].ToValue(), nil
}

func DecodeToSExp(sexp string) ([]parser.SExp, error) {
	sexps, err := parser.Parse(sexp)
	if err != nil {
		return nil, err
	}
	return sexps, nil
}

/// utilities for the result not-typed object

func ToArray(o interface{}) []interface{} {
	arr, ok := o.([]interface{})
	if !ok {
		return nil
	}
	return arr
}

func ConvertType(targetType reflect.Type, srcValue reflect.Value) (reflect.Value, error) {
	switch targetType.Kind() {
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		return ConvertArrayType(targetType, srcValue)
	case reflect.Map:
		return ConvertMapType(targetType, srcValue)
	default:
	}

	return srcValue.Convert(targetType), nil
}

func ConvertMapType(targetType reflect.Type, srcValue reflect.Value) (reflect.Value, error) {
	if srcValue.Type().Kind() != reflect.Slice {
		return reflect.ValueOf(nil), fmt.Errorf("cannot convert non-slice to map, got: [%v]",
			srcValue.Type().Kind().String())
	}

	keyType := targetType.Key()
	valType := targetType.Elem()
	retMapVal := reflect.MakeMap(targetType)

	l := srcValue.Len()
	for i := 0; i < l; i++ {
		elem := srcValue.Index(i).Interface()
		elemSlice, ok := elem.([]interface{})
		if !ok {
			return reflect.ValueOf(nil), fmt.Errorf("map element at index %d is not a slice", i)
		}

		if len(elemSlice) != 2 {
			return reflect.ValueOf(nil), fmt.Errorf("map element at index %d does not have exactly 2 elements (has %d)", i, len(elemSlice))
		}

		// Convert key and value to the target types
		keyVal, err := ConvertType(keyType, reflect.ValueOf(elemSlice[0]))
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("failed to convert key at index %d: %v", i, err)
		}

		valVal, err := ConvertType(valType, reflect.ValueOf(elemSlice[1]))
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("failed to convert value at index %d: %v", i, err)
		}

		retMapVal.SetMapIndex(keyVal, valVal)
	}

	return retMapVal, nil
}

func convertElm(lst reflect.Value, i int, elmType reflect.Type) reflect.Value {
	cv, _ := ConvertType(elmType, reflect.ValueOf(lst.Index(i).Interface()))
	return cv
}

func ConvertArrayType(targetType reflect.Type, srcValue reflect.Value) (reflect.Value, error) {
	if srcValue.Type().Kind() != reflect.Slice {
		return reflect.ValueOf(nil), fmt.Errorf("different type between target:[%v] and received:[%v]",
			targetType.Kind().String(), srcValue.Type().Kind().String())
	}
	//pp.Printf("src:%v -> %v\n", srcValue, targetType.String())
	retSliceVal := reflect.MakeSlice(targetType, srcValue.Len(), srcValue.Len())
	elmType := targetType.Elem()

	len := srcValue.Len()
	switch elmType.Kind() {
	case reflect.Int:
		for i := 0; i < len; i++ {
			cv := convertElm(srcValue, i, elmType)
			retSliceVal.Index(i).SetInt(cv.Int())
		}
	case reflect.Uint:
		for i := 0; i < len; i++ {
			cv := convertElm(srcValue, i, elmType)
			retSliceVal.Index(i).SetUint(cv.Uint())
		}
	case reflect.String:
		for i := 0; i < len; i++ {
			cv := convertElm(srcValue, i, elmType)
			retSliceVal.Index(i).SetString(cv.String())
		}
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		for i := 0; i < len; i++ {
			cv := convertElm(srcValue, i, elmType)
			retSliceVal.Index(i).SetFloat(cv.Float())
		}
	case reflect.Array:
		for i := 0; i < len; i++ {
			cv := convertElm(srcValue, i, elmType)
			retSliceVal.Index(i).Set(cv)
		}
	case reflect.Slice:
		for i := 0; i < len; i++ {
			cv := convertElm(srcValue, i, elmType)
			retSliceVal.Index(i).Set(cv)
		}
	case reflect.Map:
		for i := 0; i < len; i++ {
			cv := convertElm(srcValue, i, elmType)
			retSliceVal.Index(i).Set(cv)
		}
	case reflect.Struct:
		return retSliceVal, fmt.Errorf("converting struct not implemented")
	default:
		return retSliceVal, fmt.Errorf("converting [%v] not implemented", elmType.Kind().String())
	}

	return retSliceVal, nil
}
