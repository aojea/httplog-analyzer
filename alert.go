package main

import (
	"fmt"
	"log"
	"time"

	graphite "github.com/JensRantil/graphite-client"
)

// Alerter send an event based on metrics information
type Alerter interface {
	Alert(g *graphite.Client, threshold int) error
}

// CommonLogAlert implements the Alterter interface
type CommonLogAlert struct{}

// Alert depending of threshold
func (c CommonLogAlert) Alert(g *graphite.Client, threshold int) error {
	values, err := g.QueryIntsSince("myhost.category.value", 2*time.Minute)
	if err != nil {
		log.Println("Error querying graphite: %v", err)
	}
	fmt.Println(values)
	return nil
}
