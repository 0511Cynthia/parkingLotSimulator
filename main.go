package main

import (
	"parkingLotSimulator/stage/scene/container/models"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	estacionamiento := models.NewEstacionamiento(20) // Capacidad de 20 vehículos
	wg.Add(1)
	go estacionamiento.SimularVehiculos(&wg)
	wg.Wait()
}