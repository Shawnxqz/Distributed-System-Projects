// Implementation of a KeyValueServer. Students should write their code in this file.

package p0partA

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/cmu440/p0partA/kvstore"
	"io"
	"net"
	"strconv"
)

const MessageMaxCapacity = 500

type keyValueServer struct {
	// TODO: implement this!
	myStore       kvstore.KVStore
	listener      net.Listener
	clients       []*client
	activeNum     int
	droppedNum    int
	activeResult  chan int
	droppedResult chan int
	status        chan bool
	requests      chan *request
	connections   chan net.Conn
}

type client struct {
	conn     net.Conn
	status   chan bool
	messages chan string
}

type request struct {
	operation string
	key       string
	oldValue  string
	newValue  string
	cli       *client
}

// Create and return a new client
func newCli(newConn net.Conn) *client {
	return &client{
		conn:     newConn,
		status:   make(chan bool),
		messages: make(chan string, MessageMaxCapacity),
	}
}

// Create and return a new request
func newRequest(newOperation string, newKey string, oldValue string, newValue string, newCli *client) *request {
	return &request{
		operation: newOperation,
		key:       newKey,
		oldValue:  oldValue,
		newValue:  newValue,
		cli:       newCli,
	}
}

// New creates and returns (but does not start) a new KeyValueServer.
func New(store kvstore.KVStore) KeyValueServer {
	// TODO: implement this!
	return &keyValueServer{
		myStore:       store,
		requests:      make(chan *request),
		connections:   make(chan net.Conn),
		clients:       nil,
		activeNum:     0,
		droppedNum:    0,
		activeResult:  make(chan int),
		droppedResult: make(chan int),
		listener:      nil,
		status:        make(chan bool),
	}
}

func (kvs *keyValueServer) Start(port int) error {

	listener, err := net.Listen("tcp", "localhost:"+strconv.Itoa(port))
	if err != nil {
		fmt.Println("Listen failed, err:", err)
		//handle error
		return err
	}
	kvstore.CreateWithBackdoor()

	kvs.listener = listener

	// accepting new clients
	go func() {
		for {
			conn, err := kvs.listener.Accept()
			if err != nil {
				//fmt.Println("Accept failed, err: ", err)
				continue
			}
			kvs.connections <- conn
		}
	}()

	go mainRoutine(kvs)

	return nil
}

func (kvs *keyValueServer) Close() {

	kvs.listener.Close()
	// close the server's routine
	kvs.status <- true
}

func (kvs *keyValueServer) CountActive() int {

	kvs.requests <- newRequest("CntActive", "", "", "", nil)
	curActiveNum := <-kvs.activeResult
	return curActiveNum
}

func (kvs *keyValueServer) CountDropped() int {

	kvs.requests <- newRequest("CntDropped", "", "", "", nil)
	curDroppedNum := <-kvs.droppedResult
	return curDroppedNum
}

func mainRoutine(kvs *keyValueServer) {
	for {
		select {
		case <-kvs.status:
			return
		case request := <-kvs.requests:
			if request.operation == "Get" {
				valueList := kvs.myStore.Get(request.key)
				var message []byte
				for _, value := range valueList {
					message = append(append([]byte(request.key), ':'), value...)

				}
				newLine := []byte("\n")
				message = append(message, newLine...)

				// If slow-reading client was detected, message would be dropped
				if len(request.cli.messages) != MessageMaxCapacity {
					request.cli.messages <- string(message)
				}

			} else if request.operation == "Put" {
				kvs.myStore.Put(request.key, []byte(request.newValue))
			} else if request.operation == "Delete" {
				kvs.myStore.Delete(request.key)
			} else if request.operation == "Update" {
				kvs.myStore.Update(request.key, []byte(request.oldValue), []byte(request.newValue))
			} else if request.operation == "CntActive" {
				kvs.activeResult <- kvs.activeNum
			} else if request.operation == "CntDropped" {
				kvs.droppedResult <- kvs.droppedNum
			} else if request.operation == "Drop" {
				kvs.activeNum--
				kvs.droppedNum++
				for i, cli := range kvs.clients {
					if cli == request.cli {
						kvs.clients = append(kvs.clients[:i], kvs.clients[i+1:]...)
						break
					}
				}
			}
		case conn := <-kvs.connections:
			fmt.Println("adding new client to server")
			cli := newCli(conn)
			kvs.activeNum++
			kvs.clients = append(kvs.clients, cli)
			go readRoutine(kvs, cli)
			go writeRoutine(cli)
		}
	}
}

func readRoutine(kvs *keyValueServer, cli *client) {

	reader := bufio.NewReader(cli.conn)
	for {

		select {
		case <-cli.status:
			return
		default:

			message, err := reader.ReadBytes('\n')

			// drop client if message is empty
			if err == io.EOF {

				cli.status <- true
				dropRequest := newRequest("Drop", "", "", "", cli)
				kvs.requests <- dropRequest
				fmt.Println("Client dropped")

				return
			} else if err != nil {
				fmt.Println("Read Failed")
				return
			}
			message = bytes.Trim(message, "\n")
			// parse the message
			parsedMessage := bytes.Split(message, []byte(":"))
			operation := string(parsedMessage[0])

			if operation == "Get" {
				key := string(parsedMessage[1])
				getRequest := newRequest(operation, key, "", "", cli)
				kvs.requests <- getRequest
			} else if operation == "Put" {
				key := string(parsedMessage[1])
				value := string(parsedMessage[2])
				putRequest := newRequest(operation, key, "", value, cli)
				kvs.requests <- putRequest
			} else if operation == "Delete" {
				key := string(parsedMessage[1])
				deleteRequest := newRequest(operation, key, "", "", cli)
				kvs.requests <- deleteRequest
			} else if operation == "Update" {
				key := string(parsedMessage[1])
				oldValue := string(parsedMessage[2])
				newValue := string(parsedMessage[3])
				updateRequest := newRequest(operation, key, oldValue, newValue, cli)
				kvs.requests <- updateRequest
			}
		}
	}

}

func writeRoutine(cli *client) {
	for {
		select {
		case <-cli.status:
			return
		case message := <-cli.messages:

			cli.conn.Write([]byte(message))
		}
	}
}
