package ak

import (
	"reflect"
	"testing"
)

func Add(p1, p2 int) int {
	return p1 + p2
}

func TestRun(t *testing.T) {
	t.Log("Add typeOf = ", reflect.TypeOf(Add))
	if reflect.TypeOf(Add).Kind() == reflect.Func {
		t.Log("Add is a Func")
	}

	t.Log("Add type NumIn = ", reflect.TypeOf(Add).In(0).Name())
	t.Log("Add params is ", reflect.TypeOf(Add).Out(0))

	t.Log(reflect.ValueOf(Add).Type().Name())

	v := reflect.ValueOf(Add)

	t.Log("Add valuesOf = ", v)
	p := make([]reflect.Value, 0)
	p = append(p, reflect.ValueOf(1))
	p = append(p, reflect.ValueOf(2))

	t.Log("p", p)
	r := v.Call(p)

	for _, vv := range r {
		switch vv.Type().Kind() {
		case reflect.Int:
			t.Log("vv = ", vv.Int())
		}
	}
}
