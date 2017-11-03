package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/contribsys/faktory"
	"github.com/contribsys/faktory/util"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "./load [push|pop] [num_jobs]\n")
		os.Exit(1)
	}
	if os.Args[1] != "push" && os.Args[1] != "pop" {
		fmt.Fprintf(os.Stderr, "./load [push|pop] [num_jobs]\n")
		os.Exit(1)
	}
	count, err := strconv.ParseInt(os.Args[2], 10, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "./load [push|pop] [num_jobs]\n")
		os.Exit(1)
	}

	if os.Args[1] == "push" {
		push(int(count))
		return
	}
	pop(int(count))
}

func pop(count int) {
	time.Sleep(300 * time.Millisecond)
	client, err := faktory.Dial(faktory.DefaultServer(), "123456")
	if err != nil {
		handleError(err)
		return
	}
	defer client.Close()

	client.Beat()

	start := time.Now()
	util.Info("Popping")
	for i := 0; i < count; i++ {
		job, err := client.Fetch("default")
		if err != nil {
			handleError(err)
			return
		}
		if i%100 == 99 {
			err = client.Fail(job.Jid, os.ErrClosed, nil)
		} else {
			err = client.Ack(job.Jid)
		}
		if err != nil {
			handleError(err)
			return
		}
	}
	util.Info("Done")
	stop := time.Since(start)
	hash, err := client.Info()
	if err != nil {
		handleError(err)
		return
	}
	util.Info(hash)

	fmt.Printf("Processed %d jobs in %2f seconds, rate: %f jobs/s", count, stop.Seconds(), float64(count)/stop.Seconds())
}

func push(count int) {
	time.Sleep(300 * time.Millisecond)
	client, err := faktory.Dial(faktory.DefaultServer(), "123456")
	if err != nil {
		handleError(err)
		return
	}
	defer client.Close()

	client.Beat()

	start := time.Now()
	util.Info("Pushing")
	for i := 0; i < count; i++ {
		if err = pushJob(client, i); err != nil {
			handleError(err)
			return
		}
	}
	util.Info("Done")
	stop := time.Since(start)
	hash, err := client.Info()
	if err != nil {
		handleError(err)
		return
	}
	util.Info(hash)

	fmt.Printf("Enqueued %d jobs in %2f seconds, rate: %f jobs/s", count, stop.Seconds(), float64(count)/stop.Seconds())
}

func pushJob(client *faktory.Client, idx int) error {
	j := &faktory.Job{
		Jid:      util.RandomJid(),
		Queue:    "default",
		Type:     "SomeJob",
		Priority: uint64(rand.Intn(9) + 1),
		Args:     []interface{}{1, "string", 3},
	}
	return client.Push(j)
}

func stacks() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGQUIT)
	buf := make([]byte, 1<<20)
	for {
		<-sigs
		stacklen := runtime.Stack(buf, true)
		log.Printf("=== received SIGQUIT ===\n*** goroutine dump...\n%s\n*** end\n", buf[:stacklen])
	}
}

func handleError(err error) {
	fmt.Println(strings.Replace(err.Error(), "\n", "", -1))
}
