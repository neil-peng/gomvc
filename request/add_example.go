package request

type Add struct {
	Key    string `req:"required"`
	Value  string `req:"required"`
	Detail string `req:"optional"`
}
