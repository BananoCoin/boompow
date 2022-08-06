package models

import (
	"encoding/json"
	"strings"
	"testing"

	utils "github.com/bananocoin/boompow-next/libs/utils/testing"
)

func TestSerializeDeserializeClientRequest(t *testing.T) {
	workRequest := ClientRequest{
		RequesterEmail:       "notserialized@gmail.com",
		RequestType:          WorkGenerate,
		RequestID:            "123",
		Hash:                 "hash",
		DifficultyMultiplier: 3,
	}

	// Just want to ensure we don't leak requester emails to the client
	bytes, err := json.Marshal(workRequest)
	str := string(bytes)
	utils.AssertEqual(t, false, strings.Contains(str, "notserialized@gmail.com"))

	utils.AssertEqual(t, nil, err)

	var deserialized map[string]interface{}
	err = json.Unmarshal(bytes, &deserialized)

	utils.AssertEqual(t, nil, deserialized["requester_email"])
	utils.AssertEqual(t, "work_generate", deserialized["request_type"])
	utils.AssertEqual(t, "123", deserialized["request_id"])
	utils.AssertEqual(t, "hash", deserialized["hash"])
	utils.AssertEqual(t, float64(3), deserialized["difficulty_multiplier"])
}
