// This example shows how to create and join a nnet with multiple nodes.
// Run with default options: go run main.go
// Show usage: go run main.go -h
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/nknorg/nnet"
	"github.com/nknorg/nnet/log"
)

func create(transport string, port uint16) (*nnet.NNet, error) {
	conf := &nnet.Config{
		Port:                  port,
		Transport:             transport,
		BaseStabilizeInterval: 233 * time.Millisecond,
	}

	nn, err := nnet.NewNNet(nil, conf)
	if err != nil {
		return nil, err
	}

	err = nn.Start()
	if err != nil {
		return nil, err
	}

	return nn, nil
}

func join(transport string, localPort, seedPort uint16) (*nnet.NNet, error) {
	nn, err := create(transport, localPort)
	if err != nil {
		return nil, err
	}

	err = nn.Join(fmt.Sprintf("127.0.0.1:%d", seedPort))
	if err != nil {
		return nil, err
	}

	return nn, nil
}

func main() {
	numNodesPtr := flag.Int("n", 10, "number of nodes")
	transportPtr := flag.String("t", "tcp", "transport type, tcp or kcp")
	flag.Parse()

	if *numNodesPtr < 1 {
		log.Error("Number of nodes must be greater than 0")
		return
	}

	const createPort uint16 = 23333
	nnets := make([]*nnet.NNet, 0)

	nn, err := create(*transportPtr, createPort)
	if err != nil {
		log.Error(err)
		return
	}
	nnets = append(nnets, nn)

	for i := 0; i < *numNodesPtr-1; i++ {
		time.Sleep(112358 * time.Microsecond)

		nn, err = join(*transportPtr, createPort+uint16(i)+1, createPort)
		if err != nil {
			log.Error(err)
			continue
		}
		nnets = append(nnets, nn)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	log.Info("\nReceived an interrupt, stopping...\n")

	var wg sync.WaitGroup
	for i := 0; i < len(nnets); i++ {
		wg.Add(1)
		go func(nn *nnet.NNet) {
			nn.Stop(nil)
			wg.Done()
		}(nnets[len(nnets)-1-i])
	}
	wg.Wait()
}
