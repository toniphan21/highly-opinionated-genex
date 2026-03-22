package example

import "context"

type CreateUserInput struct {
	Name  string
	Email string
}

type createUserOp struct {
	repository Repository
	mailer     Mailer
}

func (op *createUserOp) validate(ctx context.Context, input CreateUserInput) error {
	return nil
}

func (op *createUserOp) handle(ctx context.Context, input CreateUserInput) error {
	return nil
}
