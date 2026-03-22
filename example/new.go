package example

//go:generate go run ../main.go

type Repository interface {
}

type Mailer interface {
}

type Dispatcher interface {
}

type UserStatus int

const (
	UserStatusInactive UserStatus = iota // inactive
	UserStatusActive                     // active
)

type Pill int

const (
	PillPlacebo       Pill = iota // placebo
	PillAspirin                   // aspirin
	PillIbuprofen                 // ibuprofen
	PillParacetamol               // paracetamol
	PillAcetaminophen = PillParacetamol
)
