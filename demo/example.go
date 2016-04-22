package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
)

func ExampleFunc(front string, tail string) (result string) {
	return front + "_" + tail
}

func BasicCall(args []byte) (out []byte) {

	proxy := reflect.ValueOf(ExampleFunc)

	var r []interface{}

	var buf = bytes.NewBuffer(args)
	decoder := gob.NewDecoder(buf)
	decoder.Decode(&r)

	in := make([]reflect.Value, len(r))

	for i := range in {
		in[i] = reflect.ValueOf(r[i])
	}

	result := proxy.Call(in)
	o := make([]interface{}, len(result))

	for i := range o {

		fmt.Println("----", result[i].Type().Name())

		o[i] = result[i].Interface()
	}

	var buf1 bytes.Buffer
	encoder := gob.NewEncoder(&buf1)
	encoder.Encode(o)

	return buf1.Bytes()
}

func GenRpcFunc(fptr interface{}) {

	fn := func(args []reflect.Value) []reflect.Value {

		in := make([]interface{}, len(args))

		for i := range in {
			in[i] = args[i].Interface()
		}

		var buf bytes.Buffer
		encoder := gob.NewEncoder(&buf)
		encoder.Encode(in)

		inBytes := buf.Bytes()

		outBytes := BasicCall(inBytes)

		var res []interface{}
		var outbuf = bytes.NewBuffer(outBytes)
		decoder := gob.NewDecoder(outbuf)

		decoder.Decode(&res)

		out := make([]reflect.Value, len(res))

		for i := range out {
			out[i] = reflect.ValueOf(res[i])
		}

		return out
	}

	funcV := reflect.ValueOf(fptr).Elem() //函数值
	v := reflect.MakeFunc(funcV.Type(), fn)
	funcV.Set(v)
}

func main() {

	var PresentFunc func(string, string) string

	GenRpcFunc(&PresentFunc)

	res := PresentFunc("hello", "world")

	fmt.Println(res)

}
