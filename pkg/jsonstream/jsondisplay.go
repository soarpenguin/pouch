package jsonstream

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/containerd/containerd/pkg/progress"
	"golang.org/x/crypto/ssh/terminal"
)

// bufwriter defines interface which has Write and Flush behaviors.
type bufwriter interface {
	Write([]byte) (int, error)
	Flush() error
}

// DisplayJSONMessagesToStream prints json messages to the output stream.
func DisplayJSONMessagesToStream(body io.ReadCloser) error {
	var (
		output bufwriter = bufio.NewWriter(os.Stdout)

		start      = time.Now()
		isTerminal = terminal.IsTerminal(int(os.Stdout.Fd()))
	)

	if isTerminal {
		output = progress.NewWriter(os.Stdout)
	}

	pos := make(map[string]int)
	status := []JSONMessage{}

	dec := json.NewDecoder(body)
	for {
		var (
			msg  JSONMessage
			msgs []JSONMessage
		)

		if err := dec.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		change := true
		if _, ok := pos[msg.ID]; !ok {
			status = append(status, msg)
			pos[msg.ID] = len(status) - 1
		} else {
			change = (status[pos[msg.ID]].Status != msg.Status)
			status[pos[msg.ID]] = msg
		}

		// only display the new status if the stdout is not terminal
		if !isTerminal {
			// if the status doesn't change, skip to avoid duplicate status
			if !change {
				continue
			}
			msgs = []JSONMessage{msg}
		} else {
			msgs = status
		}

		if err := displayImageReferenceProgress(output, isTerminal, msgs, start); err != nil {
			return fmt.Errorf("failed to display progress: %v", err)
		}

		if err := output.Flush(); err != nil {
			return fmt.Errorf("failed to display progress: %v", err)
		}
	}
	return nil
}

// displayImageReferenceProgress uses tabwriter to show current progress status.
func displayImageReferenceProgress(output io.Writer, isTerminal bool, msgs []JSONMessage, start time.Time) error {
	var (
		tw      = tabwriter.NewWriter(output, 1, 8, 1, ' ', 0)
		current = int64(0)
	)

	for _, msg := range msgs {
		if msg.Error != nil {
			return fmt.Errorf(msg.Error.Message)
		}

		if msg.Detail != nil {
			current += msg.Detail.Current
		}

		status := ProcessStatus(!isTerminal, msg)
		if _, err := fmt.Fprint(tw, status); err != nil {
			return err
		}
	}

	// no need to show the total information if the stdout is not terminal
	if isTerminal {
		_, err := fmt.Fprintf(tw, "elapsed: %-4.1fs\ttotal: %7.6v\t(%v)\t\n",
			time.Since(start).Seconds(),
			progress.Bytes(current),
			progress.NewBytesPerSecond(current, time.Since(start)))
		if err != nil {
			return err
		}
	}
	return tw.Flush()
}
