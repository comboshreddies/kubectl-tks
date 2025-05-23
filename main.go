package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	//  "time"

	"kubectl-tks/cmd"
)

func main() {
	var stopIt bool
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	done := make(chan bool, 1)

	intHandler := func() {
		//      ticker := time.NewTicker(time.Millisecond * 200)
		for {
			select {
			case sig := <-signals:
				if sig.String() == "quit" {
					fmt.Println("sig quit received")
				}
				if sig.String() == "interrupt" {
					fmt.Println("sig int received")
				}
				fmt.Println()
				fmt.Println(sig)
				done <- true
				//      case <- ticker.C :
				//            fmt.Println("ticker")
			}
		}
	}

	go intHandler()

	executor := func(ctx context.Context, stop *bool) {
		select {
		case <-ctx.Done():
			fmt.Println("Worker: context canceled, exiting...")
			cancel()
			return
		default:
			cmd.Execute()
			break
		}
		//   fmt.Println("execute ended")
		//   time.Sleep(3 * time.Second)
		//   fmt.Println("sleep ended")
		//   cancel()
		done <- true
	}

	stopIt = false
	go executor(ctx, &stopIt)
	<-done
	cancel()
}
