package reflect_rpc_demo

import (
	"fmt"
	"net"
	"reflect"
	"sync"
)

type Server struct {
	sync.Mutex
	network  string
	addr     string
	listener net.Listener
	running  bool

	//存放注册方法的映射
	funcs map[string]reflect.Value
}

func NewServer(network, addr string) *Server {
	RegisterType(RpcError{})
	s := new(Server)
	s.network = network
	s.addr = addr

	s.funcs = make(map[string]reflect.Value)

	return s
}

func (s *Server) Start() error {

	var err error
	s.listener, err = net.Listen(s.network, s.addr)
	if err != nil {
		return err
	}

	s.running = true

	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			continue
		}
		//分发处理
		go s.onConn(conn)
	}
	return nil
}

func (s *Server) onConn(co net.Conn) {
	c := new(conn)
	c.co = co

	defer func() {
		if e := recover(); e != nil {
			//later log
			if err, ok := e.(error); ok {
				println("recover", err.Error())
			}
		}
		c.Close()
	}()

	for {
		//读取数据
		data, err := c.ReadMessage()
		if err != nil {
			println("read error", err.Error())
			return
		}

		//数据处理
		data, err = s.handle(data)
		if err != nil {
			println("handle error ", err.Error())
			return
		}

		err = c.WriteMessage(data)
		if err != nil {
			println("write error ", err.Error())
			return
		}

	}

}

func (s *Server) Stop() error {
	s.running = false

	if s.listener != nil {
		s.listener.Close()
	}

	return nil
}

//数据处理
func (s *Server) handle(data []byte) ([]byte, error) {
	name, args, err := decodeData(data)
	if err != nil {
		return nil, err
	}

	s.Lock()
	f, ok := s.funcs[name]
	s.Unlock()
	if !ok {
		return nil, fmt.Errorf("rpc %s not registered", name)
	}

	inValues := make([]reflect.Value, len(args))

	for i := 0; i < len(args); i++ {
		if args[i] == nil {
			inValues[i] = reflect.Zero(f.Type().In(i)) //参数值零值处理
		} else {
			inValues[i] = reflect.ValueOf(args[i])
		}
	}

	out := f.Call(inValues)

	outArgs := make([]interface{}, len(out))

	for i := 0; i < len(outArgs); i++ {
		outArgs[i] = out[i].Interface()
	}

	lastParamter := out[len(out)-1].Interface()

	//检查最后的参数
	if lastParamter != nil {
		if e, ok := lastParamter.(error); ok {
			outArgs[len(out)-1] = RpcError{e.Error()}
		} else {
			return nil, fmt.Errorf("final param must be error")
		}
	}
	return encodeData(name, outArgs)
}

//RPC 方法注册
func (s *Server) Register(name string, f interface{}) (err error) {

	//异常捕获
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%s is not callable", name)
		}
	}()

	v := reflect.ValueOf(f)

	nOut := v.Type().NumOut() //输出参数的数量
	if nOut == 0 || v.Type().Out(nOut-1).Kind() != reflect.Interface {
		err = fmt.Errorf("%s return final output param must be error interface", name)
		return
	}

	_, b := v.Type().Out(nOut - 1).MethodByName("Error")
	if !b {
		err = fmt.Errorf("%s return final output param must be error interface", name)
		return
	}

	s.Lock()

	//判断方法名称是否已经被注册了
	if _, ok := s.funcs[name]; ok {
		err = fmt.Errorf("%s has registered", name)
		s.Unlock()
		return
	}

	s.funcs[name] = v
	s.Unlock()

	return
}
