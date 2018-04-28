package message

func newMessage() *EndpointMessageTemplate {
	msg, err := NewEndpointMessageTemplate(
		"123",
		"123:123",
		"456:456",
		"",
		"")
	if err != nil {
		panic(err)
	}
	return msg
}

/*func TestEndpointMessageTemplate_ParsParamsFromConfig(t *testing.T) {
	testMessage := newMessage()
	params, err := testMessage.ParsParamsFromConfig(
		"/home/bik/go/src/github.com/privatix/dappctrl/data/server.conf")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(params)
}*/
