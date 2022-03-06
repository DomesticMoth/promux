// Promux by  DomesticMoth
//
// To the extent possible under law, the person who associated CC0 with
// promux has waived all copyright and related or neighboring rights
// to promux.
//
// You should have received a copy of the CC0 legalcode along with this
// work.  If not, see <http://creativecommons.org/publicdomain/zero/1.0/>.
package main

import (
	"os"
	"io"
	"fmt"
	"net"
	"time"
	"strings"
	"io/ioutil"
)

type Target struct{
	Addr string
	Delay time.Duration
}

type Config struct{
	Listen string
	Targets []Target
}

func handle(err error){
	if err != nil {
		panic(err)
	}
}

func ReadConfig() []Config {
	b, err := ioutil.ReadFile(os.Args[1])
	handle(err)
	configs := strings.Split(string(b), "---")
	ret := []Config{}
	for _, config := range configs{
		config = strings.TrimSuffix(config, "\n")
		config = strings.TrimPrefix(config, "\n")
		rows := []string{}
		for _, row := range strings.Split(config, "\n") {
			if !strings.HasPrefix(row, "#") {
				rows = append(rows, row)
			}
		}
		listen := rows[0]
		raw_targets := rows[1:]
		targets := []Target{}
		for _, tg := range raw_targets {
			sp := strings.Split(tg, " ")
			addr := sp[0]
			delay := "100s"
			if len(sp) > 1 {
				delay = sp[1]
			}
			d, err := time.ParseDuration(delay)
			handle(err)
			targets = append(targets, Target{addr, d})
		}
		ret = append(ret, Config{listen, targets})
	}
	return ret
}

func Shift(input io.ReadCloser, output io.WriteCloser){
	defer input.Close()
	defer output.Close()
	buf := make([]byte, 1024)
	for {
		size, err := input.Read(buf)
		if err != nil { return }
		data := buf[:size]
		_, err = output.Write(data)
		if err != nil { return }
	}
}

func Accept(inp net.Conn, targets []Target) {
	for _, target := range targets {
		dialer := net.Dialer{Timeout: time.Duration(target.Delay)}
		out, err := dialer.Dial("tcp", target.Addr)
		if err != nil { continue }
		go Shift(out, inp)
		go Shift(inp, out)
		return
	}
}

func Run(conf Config){
	socket, err := net.Listen("tcp", conf.Listen)
	handle(err)
	defer socket.Close()
	for {
		conn, err := socket.Accept()
		handle(err)
		go Accept(conn, conf.Targets)
	}
}

func main(){
	whaiter := make(chan interface{})
	conf := ReadConfig()
	for _, c := range conf {
		go Run(c)
	}
	fmt.Println("Started")
	<-whaiter
}
