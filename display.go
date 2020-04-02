package main

import (
	"fmt"
	"log"
	"time"

	graphite "github.com/JensRantil/graphite-client"
)

// Displayer displays the metrics information
type Displayer interface {
	Display(g *graphite.Client) error
}

// CommonLogDisplay implements the Displayer interface
type CommonLogDisplay struct{}

// Display metrics from graphite
func (c CommonLogDisplay) Display(g *graphite.Client) error {
	values, err := g.QueryIntsSince("myhost.category.value", 10*time.Second)
	if err != nil {
		log.Println("Error querying graphite: %v", err)
	}
	fmt.Println(values)
	return nil
}
