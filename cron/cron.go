package cron

import (
	"log"

	_cron "github.com/robfig/cron"
)

type Cron struct {
	cron           *_cron.Cron
	errChan        chan error
	jobs           []job
	distLockEnable bool
}

// JobFunc function signature for job functions
type JobFunc func() error
type job struct {
	Name    string
	CronTab string
	Handler JobFunc
}

type Opts struct {
	// enable distributed lock in cluster nodes.
	// True: only one instance can execute.
	// False: multiple instances can execute, not preferable for multiple instances.
	DistLockEnable bool
}

func New(opts *Opts) *Cron {
	return &Cron{
		cron:           _cron.New(),
		errChan:        make(chan error),
		distLockEnable: opts.DistLockEnable,
	}
}

func (c *Cron) Register(name string, cronTab string, fn JobFunc) {
	c.jobs = append(c.jobs, job{Name: name, CronTab: cronTab, Handler: fn})
}

func (c *Cron) init() {
	for _, j := range c.jobs {
		c.cron.AddFunc(j.CronTab, func(j job) func() {
			return func() {
				log.Printf("Cron [%s] invoked\n", j.Name)

				defer func() {
					if err := recover(); err != nil {
						log.Println("Cron got panic")
					}
				}()

				if err := j.Handler(); err != nil {
					log.Printf("Cron got error %v\n", err)
					c.errChan <- err
				}
				log.Printf("Cron [%s] executed\n", j.Name)
			}
		}(j))
		log.Printf("Cron [%s] registered\n", j.Name)
	}
}

func (c *Cron) Run() {
	log.Println("RUNNING CRON SERVICE...")
	defer log.Println("CRON SERVICE IS RUNNING")
	c.init()
	// TODO add distributed lock
	if c.distLockEnable {
		// dist lock logic here
	} else {
		c.cron.Start()
		go func(c *Cron) {
			for {
				log.Printf("[ERR] Cron error: %v\n", <-c.errChan)
			}
		}(c)
	}
}

func (c *Cron) stop() {
	defer log.Println("Cron stopped")
	c.cron.Stop()
}
