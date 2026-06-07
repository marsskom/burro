package proxy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"gitlab.com/marsskom/burro/internal/logger"
	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/websocket"
)

func (px *Proxy) handleWebSocket(
	clientConn net.Conn,
	ctx *model.RequestContext,
	r *http.Request,
) error {
	addr := r.Host
	if !strings.Contains(addr, ":") {
		addr += ":80"
	}

	var upConn net.Conn
	var err error
	if ctx.Protocol == model.WSS {
		upConn, err = tls.Dial("tcp", addr, &tls.Config{
			InsecureSkipVerify: px.tls.Insecure,
		})
	} else {
		upConn, err = net.Dial("tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("handleWebSocket: cannot create upstream connection: %w", err)
	}
	defer upConn.Close()

	r.RequestURI = ""
	r.URL.Scheme = "http"
	r.URL.Host = r.Host

	err = r.Write(upConn)
	if err != nil {
		return fmt.Errorf("handleWebSocket: error on upgrade request: %w", err)
	}

	br := bufio.NewReader(upConn)
	resp, err := http.ReadResponse(br, r)
	if err != nil {
		return fmt.Errorf("handleWebSocket: error read response: %w", err)
	}

	br.Reset(bytes.NewReader(nil))
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		return errors.New("handleWebSocket: not websocket upgrade")
	}

	err = resp.Write(clientConn)
	if err != nil {
		return fmt.Errorf("handleWebSocket: error write a response: %w", err)
	}

	px.plugins.EmitWSOpen(ctx)

	errCh := make(chan error, 2)

	go func() {
		buf := make([]byte, 4096)
		stream := websocket.NewWSStream()

		for {
			n, err := clientConn.Read(buf)
			if err != nil {
				errCh <- err
				return
			}

			// Forwards raw bytes.
			_, werr := upConn.Write(buf[:n])
			if werr != nil {
				errCh <- werr
				return
			}

			// Writes into stream.
			stream.Write(buf[:n])

			// Parses frames.
			for {
				frame, err := stream.NextFrame()
				if err != nil {
					// Not full frame yet.
					break
				}

				opCode := model.ToWSOpCode(frame.OpCode)
				text := ""
				if opCode == model.WSText {
					text = string(frame.Payload)
				}

				err = px.plugins.EmitWSMessage(ctx, &model.WSMessage{
					Direction: model.WSClientToServer,
					OpCode:    opCode,
					Data:      frame.Payload,
					Text:      text,
					Timestamp: time.Now().UnixMilli(),
				})
				if err != nil {
					logger.Error("error on emit ws message", "err", err)
				}
			}

			// Memory cleanup.
			if stream.GetPos() > 4096 {
				stream.Compact()
			}
		}
	}()

	go func() {
		buf := make([]byte, 4096)
		stream := websocket.NewWSStream()

		for {
			n, err := upConn.Read(buf)
			if err != nil {
				errCh <- err
				return
			}

			// Forwards raw bytes.
			_, werr := clientConn.Write(buf[:n])
			if werr != nil {
				errCh <- werr
				return
			}

			// Writes into stream.
			stream.Write(buf[:n])

			// Parses frames.
			for {
				frame, err := stream.NextFrame()
				if err != nil {
					// Not full frame yet.
					break
				}

				opCode := model.ToWSOpCode(frame.OpCode)
				text := ""
				if opCode == model.WSText {
					text = string(frame.Payload)
				}

				err = px.plugins.EmitWSMessage(ctx, &model.WSMessage{
					Direction: model.WSServerToCLient,
					OpCode:    opCode,
					Data:      frame.Payload,
					Text:      text,
					Timestamp: time.Now().UnixMilli(),
				})
				if err != nil {
					logger.Error("error on emit ws message", "err", err)
				}
			}

			// Memory cleanup.
			if stream.GetPos() > 4096 {
				stream.Compact()
			}
		}
	}()

	var err1, err2 error

	for range 2 {
		err := <-errCh
		if err1 == nil {
			err1 = err
		} else {
			err2 = err
		}
	}

	px.plugins.EmitWSClose(ctx)

	if err1 != nil && !errors.Is(err1, io.EOF) {
		return fmt.Errorf("ws pipe error: %w", err1)
	}
	if err2 != nil && !errors.Is(err2, io.EOF) {
		return fmt.Errorf("ws pipe error: %w", err2)
	}

	return nil
}
