package main

import (
  "fmt"
  "bufio"
  "net"
  "os"
  "github.com/aizatto/gosmtpd"
)

func main() {
  err := start()
  if err != nil {
    fmt.Printf("%s\n", err)
    os.Exit(1)
  }
}

func start() (err error) {
  listener, err := net.Listen("tcp", ":6666")
  if err != nil {
    return
  }

  fmt.Printf("Listening on %s\n", listener.Addr())

  for {
    conn, err := listener.Accept()
    fmt.Printf("accepted\n")
    if err != nil {
      return err
    }

    client := gosmtpd.Client{
      Conn: conn,
      Bufin: bufio.NewReader(conn),
      Bufout: bufio.NewWriter(conn),
    };

    go client.Process()
  }
}
