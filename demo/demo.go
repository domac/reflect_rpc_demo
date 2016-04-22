package main

import (
	"bytes"
	"encoding/gob"
	"reflect"
)

func RemoteOnline(id uint64) (uint64, string) {
	return id, "abc"
}

//调用的重点
func CallRpc(data []byte) []byte {
	v := reflect.ValueOf(RemoteOnline)

	var r []interface{}

	var buf = bytes.NewBuffer(data)
	var dec = gob.NewDecoder(buf)

	dec.Decode(&r)

	in := make([]reflect.Value, len(r))
	for i := range in {
		in[i] = reflect.ValueOf(r[i])
	}

	out := v.Call(in)

	o := make([]interface{}, len(out))

	for i := range o {
		o[i] = out[i].Interface()
	}

	var buf1 bytes.Buffer
	var enc = gob.NewEncoder(&buf1)

	enc.Encode(o)

	return buf1.Bytes()

}

func MakeRpc(fptr interface{}) {

	f := func(in []reflect.Value) []reflect.Value {

		args := make([]interface{}, len(in)) //入参参数
		for i := range args {
			args[i] = in[i].Interface() //参数对象
		}

		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		enc.Encode(args) //把args 序列化

		b := CallRpc(buf.Bytes())

		var buf1 = bytes.NewBuffer(b)
		dec := gob.NewDecoder(buf1)

		var r []interface{}
		dec.Decode(&r)

		out := make([]reflect.Value, len(r))
		for i := range out {
			out[i] = reflect.ValueOf(r[i])
		}

		return out
	}

	fn := reflect.ValueOf(fptr).Elem()
	v := reflect.MakeFunc(fn.Type(), f)
	fn.Set(v)
}

func main() {

	var RpcOnline func(uint64) (uint64, string)

	MakeRpc(&RpcOnline)

	a, b := RpcOnline(1000)

	println(a, b)
}
