package worker

import (
	// "log"
	// "sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Job struct {
    ID int32
    Payload []byte
}

// Pool represents the worker pool structure
type Pool struct {
    WorkerCount int
    WorkerChannel chan chan Job
    JobQueue *amqp.Channel
    Stopped chan bool
}

// Worker represents the actual worker doing the job
type Worker struct{
    ID int
    JobChannel chan Job
    WorkerChannel chan chan Job // used to communicate between dispatcher and worker
    Quit chan bool
}

// NewPool returns contructs and returns new Pool object
func NewPool(workerCount int, jobQueue *amqp.Channel) Pool {
    return Pool{
        WorkerCount: workerCount,
        WorkerChannel: make(chan chan Job),
        JobQueue: jobQueue,
        Stopped: make(chan bool),
    }
}

// func(p *Pool) Run(){
//     log.Println("Spawning the workers")

//     for i := range p.WorkerCount {
//         worker := Worker{
//             ID: i+1,
//             JobChannel: make(chan Job),
//             WorkerChannel: p.WorkerChannel,
//             Quit: make(chan bool),
//         }
//         worker.start()
//     }

//     p.Allocate();
// }

// func(w *Worker)start(){

//     go func() {
//         for{
//             w.JobChannels <- w.JobChannel // when the worker is available place channel in queue
//             select {
//             case job := <-w.JobChannel: // worker has recived job
//                 w.work(job)
//             case <-w.Quit:
//                 return 
//             }
//         }
//     }()
// }