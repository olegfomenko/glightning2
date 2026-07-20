package plugin

import (
	"bufio"
	"context"
	"errors"
	"golang.org/x/exp/jsonrpc2"
	"golang.org/x/sync/errgroup"
	"os"
)

const logsPerHandlerMultiplier = 100
const serviceGoroutines = 3

func (p *Plugin) Start(ctx context.Context, in, out *os.File, extra ...func(context.Context) error) error {
	p.registerLifecycleMethods()

	inChan := make(chan []byte)
	defer close(inChan)

	outChan := make(chan []byte, p.maxConcurrentRequests)
	defer close(outChan)

	notifyChan := make(chan []byte, p.maxConcurrentRequests*logsPerHandlerMultiplier)
	p.logger = NewLogger(notifyChan)

	wg, ctx := errgroup.WithContext(ctx)
	// Set max goroutines of processRequest to outChanCapacity
	wg.SetLimit(p.maxConcurrentRequests + serviceGoroutines + len(extra))

	scanner := prepareScanner(in, p.maxIntakeBuffer)

	for _, f := range extra {
		wg.Go(func() error {
			return f(ctx)
		})
	}

	wg.Go(func() error {
		return p.read(ctx, scanner, inChan)
	})

	wg.Go(func() error {
		return p.listen(ctx, wg, inChan, outChan)
	})

	wg.Go(func() error {
		return p.write(ctx, out, outChan, notifyChan)
	})

	return wg.Wait()
}

func (p *Plugin) listen(ctx context.Context, wg *errgroup.Group, inChan <-chan []byte, outChan chan<- []byte) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case buf := <-inChan:
			wg.Go(func() error {
				return p.processRequest(ctx, buf, outChan)
			})
		}
	}
}

func (p *Plugin) read(ctx context.Context, scanner *bufio.Scanner, inChan chan<- []byte) error {
	for scanner.Scan() {
		msg := scanner.Bytes()
		buf := make([]byte, len(msg))
		copy(buf, msg)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case inChan <- buf:
		}
	}

	if scanner.Err() == nil {
		return errors.New("received EOF (unexpected)")
	}

	return scanner.Err()
}

func (p *Plugin) write(ctx context.Context, out *os.File, outChan <-chan []byte, notifyChan <-chan []byte) error {
	outWriter := bufio.NewWriter(out)

	for {
		var msg []byte
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg = <-outChan:
		case msg = <-notifyChan:
		}

		twoNewlines := []byte("\n\n")
		if _, err := outWriter.Write(append(msg, twoNewlines...)); err != nil {
			return err
		}
		if err := outWriter.Flush(); err != nil {
			return err
		}
	}
}

func (p *Plugin) processRequest(ctx context.Context, buf []byte, outChan chan<- []byte) (err error) {
	msg, err := jsonrpc2.DecodeMessage(buf)
	if err != nil {
		return err
	}

	if request := msg.(*jsonrpc2.Request); request.IsCall() {
		var response jsonrpc2.Message
		handler, ok := p.requestHandlers[request.Method]

		if !ok {
			// Requested method not found
			response, err = jsonrpc2.NewResponse(request.ID, nil, jsonrpc2.ErrMethodNotFound)
			if err != nil {
				// It should not happen
				return err
			}
		} else {
			result, rerr := handler(ctx, request.Params)
			response, err = jsonrpc2.NewResponse(request.ID, result, rerr)
			if err != nil {
				// It should not happen
				return err
			}
		}

		responseMsg, err := jsonrpc2.EncodeMessage(response)
		if err != nil {
			// It should not happen
			return err
		}

		select {
		case <-ctx.Done():
			return nil
		case outChan <- responseMsg:
		}
	} else {
		if handler, ok := p.notificationHandlers[request.Method]; ok {
			// TODO log error
			_ = handler(ctx, request.Params)
		}
	}

	return nil
}

func prepareScanner(in *os.File, maxIntakeBuffer int) *bufio.Scanner {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 1024), maxIntakeBuffer)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		for i := 0; i < len(data); i++ {
			// Check if this and next symbol (if exists) is \n
			// Then we found the end of the message
			if data[i] == '\n' && (i+1) < len(data) && data[i+1] == '\n' {
				return i + 2, data[:i], nil
			}
		}
		// This trashes anything left over in the buffer if we're at EOF, with no /n/n present
		return 0, nil, nil
	})

	return scanner
}
