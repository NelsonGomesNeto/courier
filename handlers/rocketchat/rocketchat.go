package rocketchat

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/go-errors/errors"
	"github.com/nyaruka/courier"
	"github.com/nyaruka/courier/handlers"
	"github.com/nyaruka/courier/utils"
)

var apiURL = "localhost:3000/api/v1"

func init() {
	courier.RegisterHandler(newHandler())
}

type handler struct {
	handlers.BaseHandler
}

func newHandler() courier.ChannelHandler {
	return &handler{handlers.NewBaseHandler(courier.ChannelType("RC"), "Rocket.Chat")}
}

// Initialize is called by the engine once everything is loaded
func (h *handler) Initialize(s courier.Server) error {
	h.SetServer(s)
	s.AddHandlerRoute(h, http.MethodPost, "receive", h.receiveMessage)
	return nil
}

// receiveMessage is our HTTP handler function for incoming messages

func (h *handler) receiveMessage(ctx context.Context, channel courier.Channel, w http.ResponseWriter, r *http.Request) ([]courier.Event, error) {
	return nil, nil
	// payload := &moPayload{}
	// err := handlers.DecodeAndValidateJSON(payload, r)
	// if err != nil {
	// 	return nil, handlers.WriteAndLogRequestError(ctx, h, channel, w, r, err)
	// }

	// // no message? ignore this
	// if payload.Message.MessageID == 0 {
	// 	return nil, handlers.WriteAndLogRequestIgnored(ctx, h, channel, w, r, "Ignoring request, no message")
	// }

	// // create our date from the timestamp
	// date := time.Unix(payload.Message.Date, 0).UTC()

	// // create our URN
	// urn, err := urns.NewTelegramURN(payload.Message.From.ContactID, strings.ToLower(payload.Message.From.Username))
	// if err != nil {
	// 	return nil, handlers.WriteAndLogRequestError(ctx, h, channel, w, r, err)
	// }

	// // build our name from first and last
	// name := handlers.NameFromFirstLastUsername(payload.Message.From.FirstName, payload.Message.From.LastName, payload.Message.From.Username)

	// // our text is either "text" or "caption" (or empty)
	// text := payload.Message.Text

	// // this is a start command, trigger a new conversation
	// if text == "/start" {
	// 	event := h.Backend().NewChannelEvent(channel, courier.NewConversation, urn).WithContactName(name).WithOccurredOn(date)
	// 	err = h.Backend().WriteChannelEvent(ctx, event)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	return []courier.Event{event}, courier.WriteChannelEventSuccess(ctx, w, r, event)
	// }

	// // normal message of some kind
	// if text == "" && payload.Message.Caption != "" {
	// 	text = payload.Message.Caption
	// }

	// // we had an error downloading media
	// if err != nil && text == "" {
	// 	return nil, handlers.WriteAndLogRequestIgnored(ctx, h, channel, w, r, fmt.Sprintf("unable to resolve file: %s", err.Error()))
	// }

	// // build our msg
	// msg := h.Backend().NewIncomingMsg(channel, urn, text).WithReceivedOn(date).WithExternalID(fmt.Sprintf("%d", payload.Message.MessageID)).WithContactName(name)

	// // and finally write our message
	// return handlers.WriteMsgsAndResponse(ctx, h, []courier.Msg{msg}, w, r)
}

func (h *handler) sendMsgPart(msg courier.Msg, token string, path string, form string) (string, *courier.ChannelLog, error) {
	sendURL := "localhost:3000/api/v1/method.call/sendMessage"
	// var jsonStr = []byte(`{"message":"{\"msg\":\"method\",\"method\":\"sendMessage\",\"params\":[{\"_id\":\"FsPvMfYu8YwH2cE6Y\",\"rid\":\"GENERAL\",\"msg\":\"a\"}],\"id\":\"20\"}"}`)
	var jsonStr = []byte(`{}`)
	req, _ := http.NewRequest(http.MethodPost, sendURL, bytes.NewBuffer(jsonStr))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Auth-Token", "XD2tw_bivGlsckm0Px-T41hyxXeOSTo2qX4AGjPMUPZ")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("X-User-Id", "LPCCTNijXfnTXz6pf")

	rr, err := utils.MakeHTTPRequest(req)

	// build our channel log
	log := courier.NewChannelLogFromRR("Message Sent", msg.Channel(), msg.ID(), rr).WithError("Message Send Error", err)

	// was this request successful?
	asasdfasdfjkl := rr.StatusCode
	println(asasdfasdfjkl)
	ok, err := jsonparser.GetBoolean([]byte(rr.Body), "success")
	if err != nil || !ok {
		return "", log, errors.Errorf("response not 'ok'")
	}

	// grab our message id
	externalID, err := jsonparser.GetInt([]byte(rr.Body), "result", "message_id")
	if err != nil {
		return "", log, errors.Errorf("no 'result.message_id' in response")
	}

	return strconv.FormatInt(externalID, 10), log, nil
}

// SendMsg sends the passed in message, returning any error
func (h *handler) SendMsg(ctx context.Context, msg courier.Msg) (courier.MsgStatus, error) {
	confAuth := msg.Channel().ConfigForKey(courier.ConfigAuthToken, "")
	authToken, isStr := confAuth.(string)
	if !isStr || authToken == "" {
		return nil, fmt.Errorf("invalid auth token config")
	}

	status := h.Backend().NewMsgStatusForID(msg.Channel(), msg.ID(), courier.MsgErrored)

	form := `{"message":"{\"msg\":\"method\",\"method\":\"sendMessage\",\"params\":[{\"_id\":\"FsPvMfYu8YwH2cE6Y\",\"rid\":\"GENERAL\",\"msg\":\"a\"}],\"id\":\"20\"}"}`

	externalID, log, err := h.sendMsgPart(msg, authToken, "sendMessage", form)
	status.SetExternalID(externalID)
	hasError := err != nil
	status.AddLog(log)

	if !hasError {
		status.SetStatus(courier.MsgWired)
	}

	return status, nil
}

/* Authorization example:
X-Auth-Token
X-Requested-With: XMLHttpRequest
X-User-Id
*/
/* Payload example:
{
	message: "{
		"msg": "method",
		"method": "sendMessage",
		"params": [{"_id":"AwSJtbtQqZ4cSghLm","rid":"GENERAL","msg":"hehe"}],
		"id":"47"
	}"
}
*/
type rocketChatPayload struct {
	Message struct {
		Msg    string `json:"msg"`
		Method string `json:"method"`
		Params []struct {
			ID  string `json: "_id"`
			Rid string `json: "rid"`
			Msg string `json: "msg"`
		} `json:"params"`
		ID string `json:"id"`
	} `json:"message"`
}
