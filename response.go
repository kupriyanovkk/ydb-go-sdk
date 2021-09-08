package ydb

import "github.com/ydb-platform/ydb-go-sdk/v3/internal"

type Response = internal.Response

type OpResponse = internal.OpResponse

func WrapOpResponse(resp OpResponse) Response {
	return internal.WrapOpResponse(resp)
}

type NoOpResponse = internal.NoOpResponse

func WrapNoOpResponse(resp NoOpResponse) Response {
	return internal.WrapNoOpResponse(resp)
}
