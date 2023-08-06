package handler

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
)


var chat = Chat(Id:1)

func TestParseUpdateMessageWithText(t *testing.T) {
	var msg = Message{
		Text: "Hello, world",
		Chat: chat,
	}

	var update = Update{
		UpdateId: 1,
		Message: msg,
	}

	requestBody, err := json.Marshal(update)
	if err != nil {
		t.Errorf("Failed to marshal update in json, got %s", err.Error())

	}
	
	req := httptest.NewRequest("POST", "http://myTelegramWebHookHandler.com/secretToken", bytes.NewBuffer(requestBody))

	var updateToTest, errParse = parseTelegramRequest(req)
	if errParse != nil {
		t.Errorf("Expected a <nil> error, got %s", errParse.Error())
	}
	if *updateToTest != update {
		t.Errorf("Expected update %s, got %s", update, updateToTest)
	}

}