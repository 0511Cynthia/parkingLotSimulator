package main

import (
	"fmt"
	"parkingLotSimulator/stage/scene/container/models"
	"parkingLotSimulator/ui"
	"sync"
)

func main() {
	var wg sync.WaitGroup

	// Crear el parking y el sujeto
	parking := models.NewParking(20)
	subject := parking.GetSubject()

	// Crear la vista
	view := ui.NewParkingLotView(subject)

	// Ejecutar la simulación en una goroutine
	wg.Add(1)
	go parking.SimulateCars(&wg)

	// Detener la simulación al cerrar la ventana
	view.Window.SetOnClosed(func() {
		fmt.Println("Cerrando la simulación...")
		parking.Stop() // Detener el modelo
	})

	// Ejecutar la interfaz gráfica en el hilo principal
	view.Run()

	// Esperar a que las goroutines finalicen
	wg.Wait()
}
