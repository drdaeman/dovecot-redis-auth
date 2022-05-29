package main

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"
	"strings"

	"github.com/rs/xid"
	"go.uber.org/zap"
)

type contextKey int

var loggerValue contextKey

var guid = xid.New()

type HandlerFunc func(context.Context, *Request) error

type Conn struct {
	Id      string
	Handler HandlerFunc
	Log     *zap.Logger

	conn net.Conn
}

type Request struct {
	Command byte
	Args    []string
	Log     *zap.Logger

	c *Conn
}

func (c *Conn) handleConnection() {
	defer Close("close connection", nil, c.conn.Close)
	c.Log.Info("Handling new connection")

	reader := bufio.NewReader(c.conn)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			c.Log.Info("Connection closed")
			return
		} else if err != nil {
			c.Log.Error("Error reading data", zap.Error(err))
			return
		}

		if len(line) < 1 {
			c.Log.Debug("Received an empty line")
			continue
		}
		c.Log.Debug("Received a line", zap.String("RawRequest", line))

		command := line[0]
		args := strings.Split(line[1:], "\t")
		for i := range args {
			args[i] = TabUnescape(args[i])
		}

		request := &Request{
			Command: command,
			Args:    args,
			Log:     c.Log.With(zap.ByteString("Command", []byte{command})),
			c:       c,
		}
		ctx := context.WithValue(context.TODO(), loggerValue, request.Log)
		err = c.Handler(ctx, request)
		if err != nil {
			request.Log.Error("Handler returned with an error", zap.Error(err))
			err = request.Respond(ResponseFailure, err.Error())
			if err != nil {
				request.Log.Error("Failed to respond with an error", zap.Error(err))
			}
		}
	}
}

func (r *Request) RespondEmptyLine() error {
	r.c.Log.Debug("Sending an empty line", zap.String("RawResponse", ""))
	_, err := r.c.conn.Write([]byte("\n"))
	return err
}

func (r *Request) Respond(status byte, values ...string) error {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(status)
	first := true
	for _, value := range values {
		if first {
			first = false
		} else {
			buf.WriteByte('\t')
		}
		buf.WriteString(TabEscape(value))
	}
	buf.WriteByte('\n')
	data := buf.Bytes()
	r.c.Log.Debug("Sending a line", zap.ByteString("RawResponse", data))
	_, err := r.c.conn.Write(data)
	return err
}

func (r *Request) Fatal(response error) {
	err := r.Respond(ResponseFailure, response.Error())
	if err != nil {
		r.Log.Warn("Failed to respond with an error", zap.Error(err))
	}
	err = r.c.conn.Close()
	if err != nil {
		r.Log.Error("Failed to close connection", zap.Error(err))
	}
}

func NewConn(conn net.Conn, baseLogger *zap.Logger, handlerFunc HandlerFunc) *Conn {
	remoteAddr := conn.RemoteAddr().String()
	connectionId := guid.String()

	c := &Conn{
		Id:   connectionId,
		conn: conn,
		Log: baseLogger.With(
			zap.String("RemoteAddr", remoteAddr),
			zap.String("ConnectionId", connectionId),
		),
		Handler: handlerFunc,
	}
	go c.handleConnection()
	return c
}
