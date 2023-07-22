package workerpool

import (
	"fmt"
	"testing"
)

type mockreq1 struct {
	A string
}

func (m mockreq1) HandleReq() {
	fmt.Print("this is mockreq1")
}
func (m mockreq1) WrapCommunicator(c Communication) SpecReq {
	var wrapReq = SpecReq{
		ReqHandler:    m,
		Communication: c,
	}
	return wrapReq
}

type mockreq2 struct {
	B string
}

func (m mockreq2) WrapCommunicator(c Communication) SpecReq {
	var wrapReq = SpecReq{
		ReqHandler:    m,
		Communication: c,
	}
	return wrapReq
}
func (m mockreq2) HandleReq() {

}
func TestWrapper(t *testing.T) {
	a := mockreq1{}
	c := NewCommunicator()
	wrap1 := a.WrapCommunicator(&c)
	wrap1.HandleReq()
}
