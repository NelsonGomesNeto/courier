package rocketchat

import (
	"context"
	"io/ioutil"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nyaruka/courier"
	. "github.com/nyaruka/courier/handlers"
	"github.com/nyaruka/gocommon/urns"
	"github.com/sirupsen/logrus"
)

var testChannels = []courier.Channel{
	courier.NewMockChannel("8eb23e93-5ecb-45ba-b726-3b064e0c568c", "TG", "2020", "US", map[string]interface{}{"auth_token": "a123"}),
}

var helloMsg = `{
  "update_id": 174114370,
  "message": {
	"message_id": 41,
	"from": {
		"id": 3527065,
		"first_name": "Nic",
		"last_name": "Pottier",
		"username": "nicpottier"
	},
	"chat": {
		"id": 3527065,
		"first_name": "Nic",
		"last_name": "Pottier",
		"type": "private"
	},
	"date": 1454119029,
	"text": "Hello World"
  }
}`

var startMsg = `{
    "update_id": 174114370,
    "message": {
      "message_id": 41,
      "from": {
          "id": 3527065,
          "first_name": "Nic",
          "last_name": "Pottier",
          "username": "nicpottier"
      },
      "chat": {
          "id": 3527065,
          "first_name": "Nic",
          "last_name": "Pottier",
          "type": "private"
      },
      "date": 1454119029,
      "text": "/start"
    }
  }`

var emptyMsg = `{
 	"update_id": 174114370
}`

var stickerMsg = `
{
  "update_id":174114373,
  "message":{
    "message_id":44,
    "from":{
      "id":3527065,
      "first_name":"Nic",
      "last_name":"Pottier"
    },
    "chat":{
      "id":3527065,
      "first_name":"Nic",
      "last_name":"Pottier",
      "type":"private"
    },
    "date":1454119668,
    "sticker":{
      "width":436,
      "height":512,
      "thumb":{
        "file_id":"AAQDABNW--sqAAS6easb1s1rNdJYAAIC",
        "file_size":2510,
        "width":77,
        "height":90
      },
      "file_id":"BQADAwADRQADyIsGAAHtBskMy6GoLAI",
      "file_size":38440
    }
  }
}`

var invalidFileID = `
{
  "update_id":174114373,
  "message":{
    "message_id":44,
    "from":{
      "id":3527065,
      "first_name":"Nic",
      "last_name":"Pottier"
    },
    "chat":{
      "id":3527065,
      "first_name":"Nic",
      "last_name":"Pottier",
      "type":"private"
    },
    "date":1454119668,
    "sticker":{
      "width":436,
      "height":512,
      "thumb":{
        "file_id":"invalid",
        "file_size":2510,
        "width":77,
        "height":90
      },
      "file_id":"BQADAwADRQADyIsGAAHtBskMy6GoLAI",
      "file_size":38440
    }
  }
}`

var notOkFile = `
{
  "update_id":174114373,
  "message":{
    "message_id":44,
    "from":{
      "id":3527065,
      "first_name":"Nic",
      "last_name":"Pottier"
    },
    "chat":{
      "id":3527065,
      "first_name":"Nic",
      "last_name":"Pottier",
      "type":"private"
    },
    "date":1454119668,
    "sticker":{
      "width":436,
      "height":512,
      "thumb":{
        "file_id":"notok",
        "file_size":2510,
        "width":77,
        "height":90
      },
      "file_id":"BQADAwADRQADyIsGAAHtBskMy6GoLAI",
      "file_size":38440
    }
  }
}`

var noOkFile = `
{
  "update_id":174114373,
  "message":{
    "message_id":44,
    "from":{
      "id":3527065,
      "first_name":"Nic",
      "last_name":"Pottier"
    },
    "chat":{
      "id":3527065,
      "first_name":"Nic",
      "last_name":"Pottier",
      "type":"private"
    },
    "date":1454119668,
    "sticker":{
      "width":436,
      "height":512,
      "thumb":{
        "file_id":"nook",
        "file_size":2510,
        "width":77,
        "height":90
      },
      "file_id":"BQADAwADRQADyIsGAAHtBskMy6GoLAI",
      "file_size":38440
    }
  }
}`

var noFile = `
{
  "update_id":174114373,
  "message":{
    "message_id":44,
    "from":{
      "id":3527065,
      "first_name":"Nic",
      "last_name":"Pottier"
    },
    "chat":{
      "id":3527065,
      "first_name":"Nic",
      "last_name":"Pottier",
      "type":"private"
    },
    "date":1454119668,
    "sticker":{
      "width":436,
      "height":512,
      "thumb":{
        "file_id":"nofile",
        "file_size":2510,
        "width":77,
        "height":90
      },
      "file_id":"BQADAwADRQADyIsGAAHtBskMy6GoLAI",
      "file_size":38440
    }
  }
}`

var photoMsg = `
{
    "update_id": 900946525,
    "message": {
        "message_id": 85,
        "from": {
            "id": 3527065,
            "first_name": "Nic",
            "last_name": "Pottier",
            "username": "Nicpottier"
        },
        "chat": {
            "id": 3527065,
            "first_name": "Nic",
            "last_name": "Pottier",
            "username": "Nicpottier",
            "type": "private"
        },
        "date": 1493843318,
        "photo": [
            {
                "file_id": "AgADAQADtKcxG4LRUUQSQVUjfJIiiF8G6C8ABHsRSbk65AmUi3cBAAEC",
                "file_size": 1140,
                "width": 51,
                "height": 90
            },
            {
                "file_id": "AgADAQADtKcxG4LRUUQSQVUjfJIiiF8G6C8ABNEDQTuwtue6jXcBAAEC",
                "file_size": 12138,
                "width": 180,
                "height": 320
            },
            {
                "file_id": "AgADAQADtKcxG4LRUUQSQVUjfJIiiF8G6C8ABF8Fy2sccmWmjHcBAAEC",
                "file_size": 57833,
                "width": 450,
                "height": 800
            },
            {
                "file_id": "AgADAQADtKcxG4LRUUQSQVUjfJIiiF8G6C8ABA9NJzFdXskaincBAAEC",
                "file_size": 104737,
                "width": 720,
                "height": 1280
            }
        ],
        "caption": "Photo Caption"
    }
}`

var videoMsg = `
{
    "update_id": 900946526,
    "message": {
        "message_id": 86,
        "from": {
            "id": 3527065,
            "first_name": "Nic",
            "last_name": "Pottier",
            "username": "Nicpottier"
        },
        "chat": {
            "id": 3527065,
            "first_name": "Nic",
            "last_name": "Pottier",
            "username": "Nicpottier",
            "type": "private"
        },
        "date": 1493843364,
        "video": {
            "duration": 1,
            "width": 360,
            "height": 640,
            "mime_type": "video/mp4",
            "thumb": {
                "file_id": "AAQBABP2RvcvAATGjpC2zjwhKQ8xAAIC",
                "file_size": 1770,
                "width": 50,
                "height": 90
            },
            "file_id": "BAADAQADBgADgtFRRPFTAAHxLVw76wI",
            "file_size": 257507
        }
    }
}`

var voiceMsg = `
{
    "update_id": 900946531,
    "message": {
        "message_id": 91,
        "from": {
            "id": 3527065,
            "first_name": "Nic",
            "last_name": "Pottier",
            "username": "Nicpottier"
        },
        "chat": {
            "id": 3527065,
            "first_name": "Nic",
            "last_name": "Pottier",
            "username": "Nicpottier",
            "type": "private"
        },
        "date": 1493844646,
        "voice": {
            "duration": 1,
            "mime_type": "audio/ogg",
            "file_id": "AwADAQADCQADgtFRRGn8KrC-0D_MAg",
            "file_size": 4288
        }
    }
}`

var documentMsg = `
{
    "update_id": 900946532,
    "message": {
        "message_id": 92,
        "from": {
            "id": 3527065,
            "first_name": "Nic",
            "last_name": "Pottier",
            "username": "Nicpottier"
        },
        "chat": {
            "id": 3527065,
            "first_name": "Nic",
            "last_name": "Pottier",
            "username": "Nicpottier",
            "type": "private"
        },
        "date": 1493845100,
        "document": {
            "file_name": "TabFig2015prel.xls",
            "mime_type": "application/vnd.ms-excel",
            "file_id": "BQADAQADCgADgtFRRPrv9GQ95f8eAg",
            "file_size": 4540928
        }
    }
}`

var locationMsg = `
{
    "update_id": 900946534,
    "message": {
        "message_id": 94,
        "from": {
            "id": 3527065,
            "first_name": "Nic",
            "last_name": "Pottier",
            "username": "Nicpottier"
        },
        "chat": {
            "id": 3527065,
            "first_name": "Nic",
            "last_name": "Pottier",
            "username": "Nicpottier",
            "type": "private"
        },
        "date": 1493845244,
        "location": {
            "latitude": -2.890287,
            "longitude": -79.004333
        }
    }
}`

var venueMsg = `
{
    "update_id": 900946535,
    "message": {
        "message_id": 95,
        "from": {
            "id": 3527065,
            "first_name": "Nic",
            "last_name": "Pottier",
            "username": "Nicpottier"
        },
        "chat": {
            "id": 3527065,
            "first_name": "Nic",
            "last_name": "Pottier",
            "username": "Nicpottier",
            "type": "private"
        },
        "date": 1493845520,
        "location": {
            "latitude": -2.898944,
            "longitude": -79.006835
        },
        "venue": {
            "location": {
                "latitude": -2.898944,
                "longitude": -79.006835
            },
            "title": "Cuenca",
            "address": "Provincia del Azuay",
            "foursquare_id": "4c21facd9a67a59340acdb87"
        }
    }
}`

var contactMsg = `
{
    "update_id": 900946536,
    "message": {
        "message_id": 96,
        "from": {
            "id": 3527065,
            "first_name": "Nic",
            "last_name": "Pottier",
            "username": "Nicpottier"
        },
        "chat": {
            "id": 3527065,
            "first_name": "Nic",
            "last_name": "Pottier",
            "username": "Nicpottier",
            "type": "private"
        },
        "date": 1493845755,
        "contact": {
            "phone_number": "0788531373",
            "first_name": "Adolf Taxi"
        }
    }
}`

var testCases = []ChannelHandleTestCase{
	{Label: "Receive Valid Message", URL: "/c/tg/8eb23e93-5ecb-45ba-b726-3b064e0c568c/receive/", Data: helloMsg, Status: 200, Response: "Accepted",
		Name: Sp("Nic Pottier"), Text: Sp("Hello World"), URN: Sp("telegram:3527065#nicpottier"), ExternalID: Sp("41"), Date: Tp(time.Date(2016, 1, 30, 1, 57, 9, 0, time.UTC))},
}

// func TestHandler(t *testing.T) {
// 	telegramService := buildMockTelegramService(testCases)
// 	defer telegramService.Close()

// 	RunChannelTestCases(t, testChannels, newHandler(), testCases)
// }

// setSendURL takes care of setting the send_url to our test server host
func setSendURL(s *httptest.Server, h courier.ChannelHandler, c courier.Channel, m courier.Msg) {
	apiURL = "localhost:3000/api/v1"
}

var defaultSendTestCases = []ChannelSendTestCase{
	{Label: "Plain Send",
		Text: "Simple Message", URN: "telegram:12345",
		Status: "W", ExternalID: "133",
		ResponseBody: `{ "ok": true, "result": { "message_id": 133 } }`, ResponseStatus: 200,
		PostParams: map[string]string{
			"text":         "Simple Message",
			"chat_id":      "12345",
			"reply_markup": `{"remove_keyboard":true}`,
		},
		SendPrep: setSendURL},
}

func newServer(backend courier.Backend) courier.Server {
	// for benchmarks, log to null
	logger := logrus.New()
	logger.Out = ioutil.Discard
	logrus.SetOutput(ioutil.Discard)

	config := courier.NewConfig()
	config.FacebookWebhookSecret = "fb_webhook_secret"
	config.FacebookApplicationSecret = "fb_app_secret"

	return courier.NewServerWithLogger(config, backend, logger)
}

func TestSending(t *testing.T) {
	var defaultChannel = courier.NewMockChannel("8eb23e93-5ecb-45ba-b726-3b064e0c56ab", "RC", "2020", "US",
		map[string]interface{}{courier.ConfigAuthToken: "auth_token"})

	mb := courier.NewMockBackend()
	s := newServer(mb)
	handler := newHandler()
	handler.Initialize(s)
	mb.AddChannel(defaultChannel)
	testCase := defaultSendTestCases[0]
	msg := mb.NewOutgoingMsg(defaultChannel, courier.NewMsgID(10), urns.URN(testCase.URN), testCase.Text, testCase.HighPriority, testCase.QuickReplies, testCase.Topic, testCase.ResponseToID, testCase.ResponseToExternalID)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	status, err := handler.SendMsg(ctx, msg)
	cancel()
	println(status)
	println(err)
}
