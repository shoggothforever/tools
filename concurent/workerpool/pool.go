package workerpool

import (
	"fmt"
	"time"
)

type ReplyQueue chan bool
type JobHandler func()

const (
	MAXBUFFER     int           = 2048
	WORKERTIMEOUT time.Duration = 3e3
	WORKERNUMS    int           = 1000
)

var PlatFormService *WorkerPool
var DriverService *WorkerPool

func init() {
	PlatFormService = NewWorkerPool(WORKERNUMS, WORKERTIMEOUT)
	DriverService = NewWorkerPool(WORKERNUMS, WORKERTIMEOUT)
}

var TimeOut = time.Duration(5000 * time.Millisecond)

type Req interface {
	Communication
	ReqHandler
}
type ReqHandler interface {
	HandleReq()
}
type SpecReq struct {
	ReqHandler
	Communication
}
type Communication interface {
	//Rollback()
	//exitChan 向Worker传递工作处理结束的信息,handle传递工作的处理方法
	Do(exitChan ReplyQueue, handle JobHandler)
	//向通知管道发送讯息
	Send(signal bool)
	Done()
	Result() bool
}

// 装饰着模式中的 Component
type Communicator struct {
	//客户端验证结果
	result bool
	//通知worker工作处理结果
	replyChan ReplyQueue
	//客户端通信管道，告知任务完成
	DoneChan chan struct{}
}

func NewCommunicator() Communicator {
	return Communicator{
		result:    *new(bool),
		replyChan: make(ReplyQueue, 1),
		DoneChan:  make(chan struct{}, 1),
	}
}

// 实现每个Req的接口定义
func (c *Communicator) Do(exitchan ReplyQueue, handle JobHandler) {
	ticker := time.NewTicker(TimeOut)
	defer ticker.Stop()
	//实现运行时多态
	handle()
	select {
	case val := <-c.replyChan:
		fmt.Println("HandleReq res:", val)
		close(c.replyChan)
		exitchan <- val
		return
	case <-ticker.C:
		fmt.Println("time out")
		close(c.replyChan)
		exitchan <- false
		return
	}
}
func (c *Communicator) Send(signal bool) {
	c.replyChan <- signal
	c.result = signal
	c.DoneChan <- struct{}{}
}
func (c *Communicator) Done() {
	<-c.DoneChan
}
func (c *Communicator) Result() bool {
	return c.result
}

// 定义worker，用于处理请求
type ReqQueue chan Req
type Worker struct {
	ReqChan   ReqQueue
	ReplyChan ReplyQueue
}

// 定义worker的Reqchan缓冲为1
func NewWorker() *Worker {
	return &Worker{ReqChan: make(ReqQueue, 1), ReplyChan: make(ReplyQueue, 1)}
}

type WorkerQueue chan *Worker
type WorkerPool struct {
	PoolSize   int
	TimeOut    time.Duration
	ReqChan    ReqQueue
	WorkerChan WorkerQueue
	ReplyChan  ReplyQueue
}

// 定义WorkerPool 的Reqchan 的缓冲为Worker的数量
func NewWorkerPool(size int, timeout time.Duration) *WorkerPool {
	return &WorkerPool{
		PoolSize:   size,
		TimeOut:    timeout,
		ReqChan:    make(ReqQueue, size),
		WorkerChan: make(WorkerQueue, size),
		ReplyChan:  make(ReplyQueue, 1),
	}
}
func (w *Worker) Run() {
	go func() {
		var req Req
		var ok bool
		for {
			select {
			case req, ok = <-w.ReqChan:
				{
					if !ok {
						w.ReplyChan <- false
						return
					}
					req.Do(w.ReplyChan, req.HandleReq)
				}
			}
		}
	}()
}
func (p *WorkerPool) PutReq(r Req) {
	p.ReqChan <- r
}
func (p *WorkerPool) Run() {
	fmt.Println("WorkerPool 初始化")
	for i := 0; i < p.PoolSize; i++ {
		worker := NewWorker()
		p.WorkerChan <- worker
		worker.Run()
	}
	//var cnt int64
	go func() {
		for {
			select {
			case req := <-p.ReqChan:
				{
					//fmt.Println("消费者接收到的任务编号", req.(*CountRequest).OrderInfo.Cost)
					worker := <-p.WorkerChan
					worker.ReqChan <- req
					go func(worker *Worker, p *WorkerPool) {
						res := <-worker.ReplyChan
						if res == true {
							//p.ReplyChan <- true
							p.WorkerChan <- worker
						} else {
							//工人系统异常关闭工人的发送管道,工人的线程也会随之关闭,将执行失败的任务再次送回管道（可以设定重试次数）
							close(worker.ReqChan)
							<-worker.ReplyChan
							worker = nil
							newworker := NewWorker()
							p.WorkerChan <- newworker
							newworker.Run()
						}
					}(worker, p)
				}
			}
		}
	}()
}
