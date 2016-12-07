package main

import (
    "fmt"
    //"io/ioutil"
    "log"
    "math"
    "math/rand"
    "net/rpc"
    //"os"
    "time"
)


func init() {
    rand.Seed(time.Now().UnixNano())
}

func randFileName(fileLength int) string {
    rand_str := []rune("abcdefghijklmnopqrstuvwxyz")
    ret_str := make([]rune, fileLength)
    for i:=0; i<fileLength; i++ {
        ret_str[i] = rand_str[rand.Intn(len(rand_str))]
    }
    return string(ret_str)
}


func sendMsg(msg *string, size int) {
    t0 := time.Now()

    client, err := rpc.Dial("tcp", "localhost:9001")
    if err != nil {
        log.Fatal(err)
    }

    var reply bool
    err = client.Call("Handler.Ping", msg, &reply)
    if err != nil {
        log.Fatal(err)
    }
    t1 := time.Now()
    fmt.Printf("%v  |  %d\n", t1.Sub(t0), size)
    return
}

func main() {
    //wr_string := ""

    for i:=0; i<=30; i++ {
        size_f64 := math.Pow(2, float64(i))
        size_int := int(size_f64)
        msg := randFileName(size_int)
        msg_ptr := &msg
        sendMsg(msg_ptr, size_int)
    }
    return 
}

