package main

import (
	"context"
	"log"
	"os"
)

func main() {
	p := NewHoldInvoicePlugin()
	if err := p.Start(context.Background(), os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
