package main

import (
  "fmt"
  "bufio"
  "net"
  "os"
  "strings"
  "time"
)

type Client struct {
  conn net.Conn
  bufin *bufio.Reader
  bufout *bufio.Writer
  helo string
  mail_from string
  rcpt_to string
}

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

    client := Client{
      conn: conn,
      bufin: bufio.NewReader(conn),
      bufout: bufio.NewWriter(conn),
    };

    go client.Process()
  }
}

func (c *Client) Process() {
  defer c.Close()
  c.conn.SetDeadline(time.Now().Add(time.Duration(60) * time.Second))
  c.WriteStringAndFlush("220 go-smtp-server")

  // http://tools.ietf.org/html/rfc821#page-27
  // > The first command in a session must be the HELO command.
  // > If the HELO command argument is not acceptable a 501 failure
  // > reply must be returned and the receiver-SMTP must stay in
  // > the same state.
  for {
    command, err := c.ReadCommand()

    if err != nil {
      fmt.Printf("%s\n", err.Error())
      return
    }

    if strings.Index(command, "HELO") == 0 {
      c.HandleHelo(command)
      break
    } else {
      c.WriteStringAndFlush("501 Syntax error in parameters or arguments")
    }
  }

  for {
    command, err := c.ReadCommand()

    if err != nil {
      fmt.Printf("%s\n", err.Error())
      return
    }

    switch {
    // http://tools.ietf.org/html/rfc821#page-13
    case strings.Index(command, "HELO") == 0:
      c.HandleHelo(command)

    case strings.Index(command, "MAIL FROM:") == 0:
      // callback
      if len(command) > 10 {
        c.mail_from = command[10:]
      }
      c.WriteStringAndFlush("250 OK")

    case strings.Index(command, "RCPT TO:") == 0:
      // callback
      if len(command) > 8 {
        c.rcpt_to = command[8:]
      }
      c.WriteStringAndFlush("250 OK")

    case strings.Index(command, "DATA") == 0:
      c.WriteStringAndFlush(client, "354 Start mail input; end with <CRLF>.<CRLF>")
      c.HandleData()

    case strings.Index(command, "RSET") == 0:
      c.mail_from = ""
      c.rcpt_to = ""
      c.WriteStringAndFlush("250 OK")

    case strings.Index(command, "NOOP") == 0:
      c.WriteStringAndFlush("250 OK")

    case strings.Index(command, "QUIT") == 0:
      c.WriteStringAndFlush("221 <domain> Service closing transmission channel")
      return

    default:
      c.WriteStringAndFlush("500 Syntax error, command unrecognized")
    }
  }
}

func (c *Client) Close() {
  c.conn.Close()
}

func (c *Client) HandleHelo(command string) {
  if len(command) > 5 {
    c.helo = command[5:]
  }
  c.WriteStringAndFlush("250 localhost")
}

func (c *Client) HandleData() (err error) {
  data, err := c.Read("\r\n.\r\n")
  if err != nil {
    return
  }
  // callback
}

func (c *Client) WriteStringAndFlush(input string) (err error) {
  input += "\r\n"
  fmt.Printf("S: %s", input)
  _, err := c.bufout.WriteString(input)
  if err != nil {
    return
  }
  err := c.bufout.Flush()
}

func (c *Client) ReadCommand() (request string, err error) {
  request, err = c.Read("\r\n")
  if err != nil {
    return
  }

  fmt.Printf("R: %s", request)
  request = strings.Trim(request, " \r\n")
  request = strings.ToUpper(request)
  return
}

func (c *Client) Read(suffix string) (request string, err error) {
  for err == nil {
    reply, err := c.bufin.ReadString('\n')

		if reply != "" {
      request += reply
      // check for max length
    }

    if (err != nil) {
      return request, err
    }

    if strings.HasSuffix(request, suffix) {
      return request, err
    }
  }
  return request, err
}
