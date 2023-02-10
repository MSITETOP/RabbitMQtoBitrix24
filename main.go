package main

import (
    "bytes"
    "io/ioutil"
    "net/http"
    "os"
    "log"
    "sync"
    "github.com/streadway/amqp"
    "github.com/isayme/go-amqp-reconnect/rabbitmq"
)

func main() {
    rabbitmq.Debug = true
    ampqURL := os.Getenv("AMPQURL") 
    b24restURL := os.Getenv("B24REST") 
    ampqIN := os.Getenv("AMPQIN") 
    ampqOUT := os.Getenv("AMPQOUT") 
    
    conn, err := rabbitmq.Dial(ampqURL)
    if err != nil {
        log.Panic(err)
    }

    consumeCh, err := conn.Channel()
    if err != nil {
        log.Panic(err)
    }
    
    sendCh, err := conn.Channel()
    if err != nil {
        log.Panic(err)
    }

    go func() {
        d, err := consumeCh.Consume(
            ampqIN, // queue
            "",     // consumer
            false,   // auto-ack
            false,  // exclusive
            false,  // no-local
            false,  // no-wait
            nil,    // args        
        )
        if err != nil {
            log.Panic(err)
        }

        for msg := range d {
            h := msg.Headers
            method, _ := h["method"].(string)
            token, _ := h["token"].(string)
            user, _ := h["user"].(string)
            message_id := msg.MessageId
            //content_type := msg.ContentType
            jsonData := []byte(string(msg.Body))
            routing_key := msg.ReplyTo
            
            log.Printf("msg id: %s", string(message_id))
            log.Printf("msg method: %s", string(method))
            log.Printf("msg request: %s", string(jsonData))
            log.Printf("routing_key: %s", string(routing_key))
            
            if true {           //token && method && user && message_id && jsonData
                urlAdr := b24restURL+"/"+user+"/"+token+"/"+method
                resp, err := http.Post(urlAdr, "application/json", bytes.NewBuffer(jsonData))
                if err != nil {
                    log.Fatal(err)
                }
                jsonBody, _ := ioutil.ReadAll(resp.Body)
                log.Printf("msg response: %s", string(jsonBody))

                errsend := sendCh.Publish(ampqOUT, routing_key, false, false, amqp.Publishing{
                    ContentType: "text/json",
                    MessageId:   message_id,
                    Body:        []byte(jsonBody),
                }) 
                if errsend != nil { 
                    log.Printf("Error publish: %v", errsend)
                }                
            }          
            msg.Ack(true)
        }
    }()

    wg := sync.WaitGroup{}
    wg.Add(1)

    wg.Wait()
}
