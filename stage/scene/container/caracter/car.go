package caracter

import (
	"fmt"
	"math/rand"
	"time"
)

type Car struct {
	ID int
}

// Simula el tiempo que el vehículo está estacionado antes de salir
func (v *Car) Estacionarse(duracion int) {
	fmt.Printf("Vehículo %d estacionado por %d segundos\n", v.ID, duracion)
	time.Sleep(time.Duration(duracion) * time.Second)
	fmt.Printf("Vehículo %d está saliendo\n", v.ID)
}