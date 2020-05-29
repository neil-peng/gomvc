package request

type Get struct {
	Key string `req:"required"`
}
