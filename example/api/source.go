package api

import "context"

type Request struct{}
type Response struct{}

type awesomeOp struct{}

func (op *awesomeOp) validate(ctx context.Context, request Request) (int, error) {
	return 0, nil
}

func (op *awesomeOp) handle(ctx context.Context, output int) (Response, error) {
	return Response{}, nil
}
