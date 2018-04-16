package main

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"io/ioutil"
	"log"
	"time"

	"github.com/darkdarkdragon/signalr"
	"github.com/darkdarkdragon/signalr/hubs"
)

func decodeMes(mes []byte) string {
	// log.Println(mes)
	// mesbin, er := base64.StdEncoding.DecodeString(mes)
	mesbin := make([]byte, len(mes))
	log.Printf("input byte 0 %d", mes[0])
	bytesNum, er := base64.StdEncoding.Decode(mesbin, mes)
	if er != nil {
		log.Printf("error base dec: %+v %s bytes num %d", er, er, bytesNum)
		return ""
	}
	// log.Println("base decoded:", string(mesbin))
	b := bytes.NewReader(mesbin)
	z := flate.NewReader(b)
	p, err := ioutil.ReadAll(z)
	z.Close()
	if err != nil {
		log.Printf("error decompressing: %v", err)
		return ""
	}
	return string(p)
}

func main() {
	// Prepare a SignalR client.
	c := signalr.New(
		"beta.bittrex.com",
		"1.5",
		"/signalr",
		// `[{"name":"corehub"}]`,
		`[{"name":"c2"}]`,
		nil,
	)

	// Set the user agent to one that looks like a browser.
	c.Headers["User-Agent"] = "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"

	// Send note to user about CloudFlare.
	// log.Println("Bypassing CloudFlare. This takes about 5 seconds.")

	// Define message and error handlers.
	msgHandler := func(msg signalr.Message) {
		log.Printf("%+v", msg)
		if len(msg.R) > 0 {
			log.Println(decodeMes(msg.R[1:len(msg.R)-1]))
		}
		if len(msg.M) == 0 {
			return
		}
		mesb64 := msg.M[0].A[0].(string)
		// log.Println(mesb64)
		mesbin, _ := base64.StdEncoding.DecodeString(mesb64)
		b := bytes.NewReader(mesbin)
		z := flate.NewReader(b)
		p, err := ioutil.ReadAll(z)
		z.Close()
		if err != nil {
			log.Printf("error decompressing: %v", err)
			return
		}
		log.Printf("Decoded message: %s", string(p))

	}
	panicIfErr := func(err error) {
		if err != nil {
			log.Panic(err)
		}
	}

	// Start the connection.
	err := c.Run(msgHandler, panicIfErr)
	panicIfErr(err)

	// Subscribe to the USDT-BTC feed.
	err = c.Send(hubs.ClientMsg{
		H: "c2",
		// H: "corehub",
		// M: "SubscribeToExchangeDeltas",
		M: "QueryExchangeState",
		A: []interface{}{"USDT-BTC"},
		I: 1,
	})
	panicIfErr(err)
	func() {
		time.Sleep(10 * time.Second)
		err = c.Send(hubs.ClientMsg{
			H: "c2",
			// H: "corehub",
			// M: "SubscribeToExchangeDeltas",
			M: "QueryExchangeState",
			A: []interface{}{"USDT-BTC"},
			I: 2,
		})
		panicIfErr(err)
	}()

	// Wait indefinitely.
	select {}
}
