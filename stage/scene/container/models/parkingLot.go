package models

import (
	"fmt"
	"math/rand"
	"parkingLotSimulator/stage/scene/container/caracter"
	"parkingLotSimulator/ui"
	"sync"
	"time"
)

type Estacionamiento struct {
	capacidad    int
	vehiculos    int
	cajones      []bool
	acceso       chan struct{}
	mu           sync.Mutex
	subject      *ui.Subject
}

func NewEstacionamiento(capacidad int) *Estacionamiento {
	return &Estacionamiento{
		capacidad: capacidad,
		cajones:   make([]bool, capacidad),
		acceso:    make(chan struct{}, 1), // Canal de acceso exclusivo
		subject:   &ui.Subject{},
	}
}

// Método para simular la llegada de vehículos
func (e *Estacionamiento) SimularVehiculos(wg *sync.WaitGroup) {
	defer wg.Done()
	vehiculoID := 1
	for {
		// Generar vehículos en intervalos de Poisson
		time.Sleep(time.Duration(rand.ExpFloat64()*1000) * time.Millisecond)
		vehiculo := caracter.Car{ID: vehiculoID}
		vehiculoID++
		go e.IntentarEntrar(&vehiculo)
	}
}

// Intenta ingresar un vehículo al estacionamiento
func (e *Estacionamiento) IntentarEntrar(vehiculo *caracter.Car) {
	e.acceso <- struct{}{} // Bloquea el canal para controlar la entrada/salida
	e.mu.Lock()
	if e.vehiculos >= e.capacidad {
		fmt.Printf("Vehículo %d esperando (estacionamiento lleno).\n", vehiculo.ID)
		e.mu.Unlock()
		<-e.acceso
		return
	}

	// Encuentra el primer cajón disponible y ocupa el espacio
	var cajon int
	for i := range e.cajones {
		if !e.cajones[i] {
			cajon = i
			e.cajones[i] = true
			break
		}
	}
	e.vehiculos++
	fmt.Printf("Vehículo %d ha entrado al cajón %d. Vehículos actuales: %d.\n", vehiculo.ID, cajon, e.vehiculos)

	e.subject.Notify(fmt.Sprintf("Vehículo %d ha ocupado el cajón %d", vehiculo.ID, cajon))
	e.mu.Unlock()
	<-e.acceso // Libera el acceso después de entrar

	// Tiempo aleatorio de 3 a 5 segundos estacionado
	duracion := rand.Intn(3) + 3
	vehiculo.Estacionarse(duracion)
	e.SalirVehiculo(vehiculo, cajon)
}

// Simula la salida de un vehículo y libera el cajón
func (e *Estacionamiento) SalirVehiculo(vehiculo *caracter.Car, cajon int) {
	e.acceso <- struct{}{} // Bloquea el acceso para la salida
	e.mu.Lock()
	e.cajones[cajon] = false
	e.vehiculos--
	fmt.Printf("Vehículo %d ha salido del cajón %d. Vehículos actuales: %d.\n", vehiculo.ID, cajon, e.vehiculos)

	e.subject.Notify(fmt.Sprintf("Vehículo %d ha dejado el cajón %d", vehiculo.ID, cajon))
	e.mu.Unlock()
	<-e.acceso // Libera el acceso después de salir
}