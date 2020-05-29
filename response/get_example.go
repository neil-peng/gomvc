package response

type Get struct {
	Key    string `req:"optional"`
	Value  string `req:"required"`
	Detail string `req:"optional"`
}
