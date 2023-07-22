package mq

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/cmd/hz/util/logs"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"sync"
)

type RabbitConfig struct {
	Mqurl            string
	DeadName         string
	DeadQueueName    string
	DeadExchangeName string
	DeadRouting      string
	DeadExchangeType string
	BaseName         string
	BaseExchangeName string
	BaseQueueName    string
}

const (
	mqurl                   = "amqp://dsm:wusa@localhost:5672/"
	deadName                = "dead"
	deadqueuename           = "dlx_queue"
	deadexchangename        = "dlx_exchange"
	deadrouting             = "dream.dead"
	Topic                   = "topic"
	baseName                = "dsm"
	baseexchangename        = "topic_logs"
	basequeuename           = ""
	Publishdeadexchange     = "x-dead-letter-exchange"
	Publishdeadrouter       = "x-dead-letter-routing-key"
	Xmessagettl         int = 1000
	Xmaxlength          int = 1000
)

var RabbitCfg RabbitConfig
var smqs sync.Map

// var rq RabbitMQ
var once sync.Once

func init() {
	smqs = sync.Map{}
	RabbitCfg.Mqurl = mqurl
	RabbitCfg.DeadName = deadName
	RabbitCfg.DeadQueueName = deadqueuename
	RabbitCfg.DeadExchangeName = deadexchangename
	RabbitCfg.DeadExchangeType = Topic
	RabbitCfg.DeadRouting = deadrouting
	RabbitCfg.BaseName = baseName
	RabbitCfg.BaseQueueName = basequeuename
	RabbitCfg.BaseExchangeName = baseexchangename

}

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	// 连接名称
	Name string
	// 交换机
	ExchangeName string
	//
	ExchangeType string
	// routing Key
	RoutingKey string
	//MQ链接字符串
	Mqurl string
}

func (rq *RabbitMQ) setMQAtt(name, ename, etype, rkey, mqurl string) {
	rq.Name = name
	rq.ExchangeName = ename
	rq.ExchangeType = etype
	rq.RoutingKey = rkey
	rq.Mqurl = mqurl
}
func NewRabbit(name, rkey, mqurl string) (*RabbitMQ, error) {
	var rq RabbitMQ
	var err error
	irq, ok := smqs.Load(name)
	if !ok {
		rq.setMQAtt(name, "", "", rkey, mqurl)
		rq.Conn, err = amqp.Dial(mqurl)
		if err != nil {
			//mes.FailOnError(err, "与RabbitMQ建立连接失败")
			return nil, err
		}
		rq.Channel, err = rq.Conn.Channel()
		if err != nil {
			//mes.FailOnError(err, "从连接中获取管道失败")
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		smqs.Store(rq.Name, rq)
	} else {
		rq = irq.(RabbitMQ)
		fmt.Println("已经存储了该连接", rq)
	}

	return &rq, nil
}
func NewRabbitWithExchange(name, ename, etype, rkey, mqurl string) (*RabbitMQ, error) {
	var rq RabbitMQ
	var err error
	irq, ok := smqs.Load(name)
	fmt.Println(irq, ok)
	if !ok {
		rq.setMQAtt(name, ename, etype, rkey, mqurl)
		rq.Conn, err = amqp.Dial(mqurl)
		if err != nil {
			//mes.FailOnError(err, "与RabbitMQ建立连接失败")
			return nil, err
		}
		rq.Channel, err = rq.Conn.Channel()
		if err != nil {
			//mes.FailOnError(err, "从连接中获取管道失败")
			return nil, err
		}
		err = rq.Channel.ExchangeDeclare(ename, etype, true, true, false, false, nil)
		if err != nil {
			//mes.FailOnError(err, "创建Exchanger失败")
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		smqs.Store(name, rq)
		fmt.Println("存储rabbitmq连接成功", name, rq)
	} else {
		rq = irq.(RabbitMQ)
		fmt.Println("已经存储了该连接", rq)
	}
	return &rq, nil
}
func (rq *RabbitMQ) ReleaseMQ() {
	rq.Conn.Close()
	rq.Channel.Close()
}
func GetRabbitMQ(name string) RabbitMQ {
	irq, ok := smqs.Load(name)
	if !ok {
		logs.Info("获取连接信息失败，连接名%s", name)
		return RabbitMQ{}
	}
	return irq.(RabbitMQ)
}

func (rq *RabbitMQ) PublishMessage(ctx context.Context, ename, rkey string, msg amqp.Publishing) {
	err := rq.Channel.PublishWithContext(ctx, ename, rkey, true, true, msg)
	if err != nil {
		return
	}
}

// 创建死信队列
func (rq *RabbitMQ) CreateDeadQueue(dqname, dqexchange, dqrouting string, route bool) (*amqp.Queue, error) {
	//fmt.Println(rq)
	var deadq amqp.Queue
	var err error
	if route {
		deadq, err = rq.Channel.QueueDeclare(dqname, true, true, false, false, amqp.Table{
			Publishdeadexchange: dqexchange, //"dlx_exchange",
			Publishdeadrouter:   dqrouting,  //"dlx_route",
			//"x-message-ttl":     Xmessagettl, // 过期时间 死信队列就不要配置过期时间了
			"x-max-length": Xmaxlength}) // 死信队列配置参数)
	} else {
		deadq, err = rq.Channel.QueueDeclare(dqname, true, true, false, false, amqp.Table{
			Publishdeadexchange: dqexchange, //"dlx_exchange",
			//"x-message-ttl":     Xmessagettl, //过期时间 死信队列就不要配置过期时间了
			"x-max-length": Xmaxlength}) // 死信队列配置参数)
	}
	err = rq.Channel.QueueBind(dqname, dqrouting, dqexchange, false, nil)
	//mes.FailOnError(err, "绑定死信队列失败")
	return &deadq, err
}

func NewDeadQueue() {
	rbc := RabbitCfg
	rq, err := NewRabbitWithExchange(rbc.DeadName, rbc.DeadExchangeName, Topic, rbc.DeadRouting, rbc.Mqurl)
	if err != nil {
		//mes.FailOnError(err, "创建死信连接失败")
	}
	defer rq.ReleaseMQ()
	//声明死信队列
	q, err := rq.CreateDeadQueue(rbc.DeadQueueName, rbc.DeadExchangeName, rbc.DeadRouting, true)
	//mes.FailOnError(err, "创建死信队列失败")
	//消费死信队列中的信息
	msgs, err := rq.Channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	//mes.FailOnError(err, "Failed to register a consumer")
	var forever chan struct{}
	log.Println("[x] deadmq have declared and is waiting for dead messages")
	go func() {
		//业务逻辑写在这里，可以将任务丢给线程池处理
		for d := range msgs {
			//time.Sleep(5 * time.Second)
			log.Printf("有消息过期了，消息来源:%s,消息内容: %s,", d.Exchange, d.Body)
		}
	}()
	<-forever
}
