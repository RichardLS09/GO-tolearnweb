Queue
====
队列模块实现了一个非阻塞异步队列生产消费机制，对于一个队列，需要分别指定他的Producer和Consumer。
Producer和Consumer要实现对应的接口方法`Push`和`Pop`。
可以通过下面的方式初始化一个队列：
```
// import package queue

// a simple implement struct
type SimpleImpl struct{}

// implement Producer
func (p SimpleImpl) Push(q *queue.Queue, task string) {
	for i := 0; i < 5+rand.Intn(5); i++ {
		qItem := q.Push(task, i, "item data")
		log.Info("push", qItem)
	}
	time.Sleep(time.Duration(rand.Intn(2000))*time.Millisecond + 7*time.Second)
}

// implement Consumer
func (c SimpleImpl) Pop(q *queue.Queue, task string) {
	qItem := q.Pop()
	log.Info("pop", qItem)
	time.Sleep(time.Duration(rand.Intn(2000))*time.Millisecond + time.Second)
}

impl := SimpleImpl{}

// regist impl as producer because of it implement Producer interface
// regist impl as consumer because of it implement Consumer interface
// regist both producer and consumer at the same time
aQueue := queue.RegistTask("task_name", 10, impl, impl)
// here we use the task name as queue name
// nil producer or consumer will not be regist on queue in this method

// get a queue by name
targetQueue, ok := queue.GetQueue("task_name")
// aQueue and targetQueue is the same point

otherQueue := queue.CreateQueue("queue_name", 10)

// only regist producer
otherQueue.RegistProducer("producer_name", impl)

// only regist consumer
otherQueue.RegistConsumer("consumer_name", impl)

// here run all producer and consumer
// while running
// all producer.Push and consumer.Pop method will be call in loop
// if queue is empty, Pop will be block
go queue.Run()
```
