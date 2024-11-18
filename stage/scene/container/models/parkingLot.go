package models

import (
	"fmt"
	"math/rand"
	"parkingLotSimulator/stage/scene/container/caracter"
	"parkingLotSimulator/ui"
	"sync"
	"time"
)

type Parking struct {
	capacidad  int
	cars       int
	cajones    []bool
	colaEspera []int // Cola de vehículos en espera
	acceso     chan struct{}
	mu         sync.Mutex
	subject    *ui.Subject
	quit       chan struct{}
}

func NewParking(capacidad int) *Parking {
	return &Parking{
		capacidad:  capacidad,
		cajones:    make([]bool, capacidad),
		colaEspera: make([]int, 0),
		acceso:     make(chan struct{}, 1),
		subject:    &ui.Subject{},
		quit:       make(chan struct{}),
	}
}

func (e *Parking) SimulateCars(wg *sync.WaitGroup) {
	defer wg.Done()
	carID := 1

	for {
		select {
		case <-e.quit:
			fmt.Println("Deteniendo la simulación de vehículos...")
			return
		default:
			// Generar vehículos en intervalos de Poisson
			select {
			case <-time.After(time.Duration(rand.ExpFloat64()*1000) * time.Millisecond):
				car := caracter.Car{ID: carID}
				carID++
				go e.tryEnter(&car)
			case <-e.quit:
				fmt.Println("Cancelando la generación de vehículos.")
				return
			}
		}
	}
}

func (e *Parking) tryEnter(car *caracter.Car) {
	select {
	case e.acceso <- struct{}{}:
		e.mu.Lock()
		if e.cars >= e.capacidad {
			// Agregar a la cola de espera
			e.colaEspera = append(e.colaEspera, car.ID)
			fmt.Printf("Vehículo %d esperando (Estacionamiento lleno). Tamaño de cola: %d\n",
				car.ID, len(e.colaEspera))
			e.subject.Notify(fmt.Sprintf("Vehículo %d esperando en cola", car.ID))
			e.mu.Unlock()
			<-e.acceso
			return
		}

		// Si el vehículo estaba en cola, removerlo
		e.removerDeCola(car.ID)

		e.subject.Notify(fmt.Sprintf("Vehículo %d entrando", car.ID))

		// Encuentra el primer cajón disponible y ocupa el espacio
		var cajon int
		for i := range e.cajones {
			if !e.cajones[i] {
				cajon = i
				e.cajones[i] = true
				break
			}
		}
		e.cars++
		fmt.Printf("Vehículo %d ha entrado al cajón %d. Vehículos actuales: %d. Cola: %d\n",
			car.ID, cajon, e.cars, len(e.colaEspera))

		e.subject.Notify(fmt.Sprintf("Vehículo %d ha ocupado el cajón %d", car.ID, cajon))
		e.mu.Unlock()
		<-e.acceso

		// Tiempo aleatorio de 3 a 5 segundos estacionado
		duracion := rand.Intn(3) + 3
		select {
		case <-time.After(time.Duration(duracion) * time.Second):
			car.Park(duracion)
			e.carLeave(car, cajon)
		case <-e.quit:
			fmt.Printf("Vehículo %d canceló el Estacionamiento.\n", car.ID)
			return
		}
	case <-e.quit:
		fmt.Printf("Vehículo %d no pudo entrar porque se detuvo el sistema.\n", car.ID)
		return
	}
}

func (e *Parking) removerDeCola(carID int) {
	for i, id := range e.colaEspera {
		if id == carID {
			// Remover el elemento usando slice
			e.colaEspera = append(e.colaEspera[:i], e.colaEspera[i+1:]...)
			e.subject.Notify(fmt.Sprintf("Vehículo %d removido de cola", carID))
			break
		}
	}
}

func (e *Parking) carLeave(car *caracter.Car, cajon int) {
	select {
	case e.acceso <- struct{}{}:
		e.mu.Lock()
		e.cajones[cajon] = false
		e.cars--
		fmt.Printf("Vehículo %d ha salido del cajón %d. Vehículos actuales: %d. Cola: %d\n",
			car.ID, cajon, e.cars, len(e.colaEspera))
		e.subject.Notify(fmt.Sprintf("Vehículo %d ha dejado el cajón %d", car.ID, cajon))

		// Si hay vehículos en cola, intentar que entre el primero
		if len(e.colaEspera) > 0 {
			nextID := e.colaEspera[0]
			nextcar := &caracter.Car{ID: nextID}
			go e.tryEnter(nextcar)
		}

		e.mu.Unlock()
		<-e.acceso
	case <-e.quit:
		fmt.Printf("Vehículo %d canceló la salida porque se detuvo el sistema.\n", car.ID)
		return
	}
}

// Devuelve el subject para permitir el registro de observadores
func (e *Parking) GetSubject() *ui.Subject {
	return e.subject
}

func (e *Parking) Stop() {
	close(e.quit)
}
