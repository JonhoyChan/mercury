package service

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"
	"outgoing/x/types"
	"testing"
)

func TestPushMessageRequest(t *testing.T) {
	str := `{"mid": "4", "message_type": "group", "receiver": "gidqFRCSA2eLeI", "content_type": "text", "body": {"content": "Hello, World!"}}`
	req := PushMessageRequest{
		MessageType: types.MessageTypeGroup,
		Receiver:    "gidqFRCSA2eLeI",
		ContentType: types.ContentTypeText,
		Body:        []byte(`{"content": "Hello, World!"}`),
	}
	err := jsoniter.UnmarshalFromString(str, &req)
	require.NoError(t, err)
	require.True(t, req.Validate())

	req.Body = []byte(`{"contents": "Hello, World!"}`)
	require.False(t, req.Validate())

	req.ContentType = types.ContentTypeImage
	req.Body = []byte(`{"file_stat": {"filename": "v2-f7ea6b00ebcfbd1b774434bf7e839ac6.jpg","size": 115360,"width": 1253,"height": 1253},"hash": "bafykbzacedjkodrxars66qrsonrca7y6advofhrqfdtuxpkksvofu2l6slwjo"}`)
	require.True(t, req.Validate())
	req.Body = []byte(`{"content": "Hello, World!"}`)
	require.False(t, req.Validate())
}
