package models

import (
	"encoding/json"
	"testing"

	utils "github.com/bananocoin/boompow/libs/utils/testing"
)

func TestSerializeDeserializeWorkResponse(t *testing.T) {
	workResponse := ClientWorkResponse{
		RequestID: "123",
		Hash:      "hash",
		Result:    "3",
	}

	bytes, err := json.Marshal(workResponse)

	utils.AssertEqual(t, nil, err)

	var deserialized map[string]interface{}
	err = json.Unmarshal(bytes, &deserialized)

	utils.AssertEqual(t, "123", deserialized["request_id"])
	utils.AssertEqual(t, "hash", deserialized["hash"])
	utils.AssertEqual(t, "3", deserialized["result"])
}
