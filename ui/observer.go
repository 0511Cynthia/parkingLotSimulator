package ui

type Observer interface {
	Update(data interface{}) // Método para notificar cambios
}

type Subject struct {
	observers []Observer
}

func (s *Subject) Register(observer Observer) {
	s.observers = append(s.observers, observer)
}

func (s *Subject) Notify(data interface{}) {
	for _, observer := range s.observers {
		observer.Update(data)
	}
}