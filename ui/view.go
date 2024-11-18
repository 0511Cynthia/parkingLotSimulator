package ui

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

type ParkingLotView struct {
	Window         fyne.Window
	mainContainer  *fyne.Container
	cars           map[int]*canvas.Image
	rowCars        map[int]*canvas.Image
	mutex          sync.Mutex
	updateChan     chan interface{}
	quit           chan struct{}
	carEnterAsset  fyne.Resource
	carExitAsset   fyne.Resource
	carParkedAsset fyne.Resource
	app            fyne.App
	lastQueuePos   int
	animationWG    sync.WaitGroup // Para sincronizar animaciones
}

func NewParkingLotView(subject *Subject) *ParkingLotView {
	a := app.New()
	w := a.NewWindow("Simulador de Estacionamiento")
	w.Resize(fyne.NewSize(800, 500))
	w.SetFixedSize(true)

	view := &ParkingLotView{
		Window:        w,
		app:           a,
		mainContainer: container.NewWithoutLayout(),
		cars:          make(map[int]*canvas.Image),
		rowCars:       make(map[int]*canvas.Image),
		updateChan:    make(chan interface{}, 100),
		quit:          make(chan struct{}),
		lastQueuePos:  0,
	}

	view.loadAssets()

	bgResource := view.loadBackgroundResource()
	bg := canvas.NewImageFromResource(bgResource)
	bg.Resize(fyne.NewSize(800, 500))
	bg.Move(fyne.NewPos(0, 0))

	view.mainContainer.Add(bg)
	w.SetContent(view.mainContainer)

	subject.Register(view)
	go view.processUpdates()

	return view
}

func (view *ParkingLotView) loadAssets() {
	view.carEnterAsset = fyne.NewStaticResource("car_enter.png", view.loadAssetFile("car_enter.png"))
	view.carExitAsset = fyne.NewStaticResource("car_exit.png", view.loadAssetFile("car_exit.png"))
	view.carParkedAsset = fyne.NewStaticResource("car.png", view.loadAssetFile("car.png"))
}

func (view *ParkingLotView) loadAssetFile(filename string) []byte {
	fullPath := fmt.Sprintf("assets/%s", filename)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		fmt.Printf("Error loading asset %s: %v\n", filename, err)
		return []byte{} // Retornar un byte array vacío en caso de error
	}
	return data
}

func (view *ParkingLotView) loadBackgroundResource() fyne.Resource {
	bgResource, err := fyne.LoadResourceFromPath("assets/bg.png")
	if err != nil {
		fmt.Printf("Error loading background: %v\n", err)
		return fyne.NewStaticResource("bg.png", []byte{})
	}
	return bgResource
}

func (view *ParkingLotView) Update(data interface{}) {
	select {
	case view.updateChan <- data:
		// Mensaje enviado exitosamente
	default:
		fmt.Println("Canal de actualizaciones lleno, descartando mensaje")
	}
}

func (view *ParkingLotView) processUpdates() {
	for {
		select {
		case msg := <-view.updateChan:
			fmt.Printf("channel update message: %s \n", msg)
			if strMsg, ok := msg.(string); ok {
				view.handleUpdate(strMsg)
			}
		case <-view.quit:
			return
		}
	}
}

func (view *ParkingLotView) handleUpdate(msg string) {
	// Esperar a que terminen las animaciones en curso
	view.animationWG.Wait()

	switch {
	case strings.Contains(msg, "esperando en cola"):
		view.handleWaitingCar(msg)
	case strings.Contains(msg, "removido de cola"):
		view.handleRemovedFromQueue(msg)
	case strings.Contains(msg, "entrando"):
		view.handleEnteringCar(msg)
	case strings.Contains(msg, "ocupado"):
		view.handleParkingCar(msg)
	case strings.Contains(msg, "dejado"):
		view.handleExitingCar(msg)
	}
}

func (view *ParkingLotView) handleRemovedFromQueue(msg string) {
	var carID int
	fmt.Sscanf(msg, "Vehículo %d removido de cola", &carID)

	view.mutex.Lock()
	if queuedAuto, exists := view.rowCars[carID]; exists {
		view.mainContainer.Remove(queuedAuto)
		delete(view.rowCars, carID)
		view.lastQueuePos--
		view.reajustarCola()
	}
	view.mutex.Unlock()
	view.mainContainer.Refresh()
}

func (view *ParkingLotView) handleWaitingCar(msg string) {
	var carID int
	fmt.Sscanf(msg, "Vehículo %d esperando en cola", &carID)

	view.mutex.Lock()
	defer view.mutex.Unlock()

	// Verificar si el auto ya está en cola
	if _, exists := view.rowCars[carID]; exists {
		return
	}

	queueX := float32(0)
	queueY := float32(450 - (view.lastQueuePos * 60))

	auto := canvas.NewImageFromResource(view.carEnterAsset)
	auto.Resize(fyne.NewSize(50, 55))
	auto.Move(fyne.NewPos(queueX, queueY))

	view.rowCars[carID] = auto
	view.mainContainer.Add(auto)
	view.lastQueuePos++

	view.mainContainer.Refresh()
}

func (view *ParkingLotView) reajustarCola() {
	newY := float32(450)
	sortedIDs := make([]int, 0, len(view.rowCars))
	for id := range view.rowCars {
		sortedIDs = append(sortedIDs, id)
	}

	// Ordenar IDs para mantener el orden de la cola
	for i := 0; i < len(sortedIDs)-1; i++ {
		for j := 0; j < len(sortedIDs)-i-1; j++ {
			if sortedIDs[j] > sortedIDs[j+1] {
				sortedIDs[j], sortedIDs[j+1] = sortedIDs[j+1], sortedIDs[j]
			}
		}
	}

	for _, id := range sortedIDs {
		if auto, exists := view.rowCars[id]; exists {
			auto.Move(fyne.NewPos(0, newY))
			newY -= 60
			auto.Refresh()
		}
	}
}

func (view *ParkingLotView) handleEnteringCar(msg string) {
	var carID int
	fmt.Sscanf(msg, "Vehículo %d entrando", &carID)

	view.mutex.Lock()
	auto := canvas.NewImageFromResource(view.carEnterAsset)
	//auto.Resize(fyne.NewSize(50, 50))
	auto.Resize(fyne.NewSize(55, 50))
	auto.Move(fyne.NewPos(50, 400))

	view.cars[carID] = auto
	view.mainContainer.Add(auto)
	view.mutex.Unlock()

	view.mainContainer.Refresh()
	view.animationWG.Add(1)
	go func() {
		view.animateMove(carID, 50, 400, 300, 200)
		view.animationWG.Done()
	}()
}
func (view *ParkingLotView) handleParkingCar(msg string) {
	var carID, cajon int
	fmt.Sscanf(msg, "Vehículo %d ha ocupado el cajón %d", &carID, &cajon)

	view.mutex.Lock()
	auto, exists := view.cars[carID]
	view.mutex.Unlock()

	if exists {
		targetX := float32(125 + (cajon%5)*95)
		targetY := float32(80 + (cajon/5)*75)

		view.animationWG.Add(1)
		go func() {
			view.animateMove(carID, auto.Position().X, auto.Position().Y, targetX, targetY)

			view.mutex.Lock()
			if auto, exists := view.cars[carID]; exists {
				auto.Resource = view.carParkedAsset
				auto.Refresh()
			}
			view.mutex.Unlock()

			view.animationWG.Done()
		}()
	}
}

func (view *ParkingLotView) handleExitingCar(msg string) {
	var carID int
	fmt.Sscanf(msg, "Vehículo %d ha dejado el cajón %d", &carID)

	view.mutex.Lock()
	auto, exists := view.cars[carID]
	if !exists {
		view.mutex.Unlock()
		return
	}

	auto.Resource = view.carExitAsset
	auto.Refresh()

	currentX := auto.Position().X
	currentY := auto.Position().Y
	view.mutex.Unlock()

	view.animationWG.Add(1)
	go func() {
		view.animateMove(carID, currentX, currentY, 50, 400)

		time.Sleep(500 * time.Millisecond)

		view.mutex.Lock()
		if auto, exists := view.cars[carID]; exists {
			view.mainContainer.Remove(auto)
			delete(view.cars, carID)
		}
		view.mutex.Unlock()

		view.animationWG.Done()
	}()
}

func (view *ParkingLotView) animateMove(carID int, fromX, fromY, toX, toY float32) {
	steps := 30
	deltaX := (toX - fromX) / float32(steps)
	deltaY := (toY - fromY) / float32(steps)

	for i := 0; i <= steps; i++ {
		view.mutex.Lock()
		if auto, exists := view.cars[carID]; exists {
			newX := fromX + deltaX*float32(i)
			newY := fromY + deltaY*float32(i)
			auto.Move(fyne.NewPos(newX, newY))
			auto.Refresh()
		}
		view.mutex.Unlock()
		time.Sleep(13 * time.Millisecond)
	}
}

func (view *ParkingLotView) Run() {
	view.Window.ShowAndRun()
}
