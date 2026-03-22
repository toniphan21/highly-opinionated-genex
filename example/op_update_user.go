package example

import "context"

type UpdateUserInput struct {
	Name  string
	Email string
}

type updateUserOp struct {
	repository Repository
	mailer     Mailer
}

func (op *updateUserOp) validate(ctx context.Context, input UpdateUserInput) (int, error) {
	id := 0 // find user by email, passed to the downstream
	return id, nil
}

func (op *updateUserOp) authorize(input UpdateUserInput, id int) error {
	return nil
}

func (op *updateUserOp) handle(ctx context.Context, input UpdateUserInput) error {
	return nil
}
