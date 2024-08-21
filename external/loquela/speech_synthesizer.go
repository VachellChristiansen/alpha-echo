package external_loquela

import (
	"alpha-echo/models"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

const (
	STATUS_FIRST_FRAME    = 0
	STATUS_CONTINUE_FRAME = 1
	STATUS_LAST_FRAME     = 2
)

type SpeechSynthesizer interface {
	GetProperty(prop string) interface{}
	BuildRequest(text string)
	BuildFile(vocabulary *models.LoquelaVocabulary)
	BuildUrl()
	Synthesize(specifications string) (err error)
}

type SpeechSynthesizerImpl struct {
	Language   string
	Gender     string
	Audio      int
	Vocabulary *models.LoquelaVocabulary
	Request    *WebsocketRequest
	URL        string
	logger     map[string]*log.Logger
}

type WebsocketRequest struct {
	AppID        string
	APIKey       string
	APISecret    string
	Text         string
	CommonArgs   map[string]string
	BusinessArgs map[string]interface{}
	Data         map[string]interface{}
}

func NewSpeechSynthesizer(language, gender string, audio int, logger map[string]*log.Logger) SpeechSynthesizer {
	return &SpeechSynthesizerImpl{
		Language: language,
		Gender:   gender,
		Audio:    audio,
		logger:   logger,
	}
}

func (e *SpeechSynthesizerImpl) GetProperty(prop string) interface{} {
	switch prop {
	case "URL":
		return e.URL
	case "Request":
		return *e.Request
	default:
		return nil
	}
}

func (e *SpeechSynthesizerImpl) BuildRequest(text string) {
	textEncoded := base64.StdEncoding.EncodeToString([]byte(text))
	e.Request = &WebsocketRequest{
		AppID:     os.Getenv("XFYUN_TTS_APP_ID"),
		APIKey:    os.Getenv("XFYUN_TTS_API_KEY"),
		APISecret: os.Getenv("XFYUN_TTS_API_SECRET"),
		Text:      text,
		CommonArgs: map[string]string{
			"app_id": os.Getenv("XFYUN_TTS_APP_ID"),
		},
		BusinessArgs: map[string]interface{}{
			"aue": "lame",
			"auf": "audio/L16;rate=16000",
			"vcn": "x_xiaolin",
			"tte": "utf8",
			"sfl": 1,
		},
		Data: map[string]interface{}{
			"status": STATUS_LAST_FRAME,
			"text":   textEncoded,
		},
	}
}

func (e *SpeechSynthesizerImpl) BuildUrl() {
	date := time.Now().UTC().Format(time.RFC1123)

	signatureOrigin := "host: " + os.Getenv("XFYUN_TTS_BASE_HOST") + "\n" +
		"date: " + date + "\n" +
		"GET " + "/v2/tts " + "HTTP/1.1"

	hashedSecret := hmac.New(sha256.New, []byte(e.Request.APISecret))
	hashedSecret.Write([]byte(signatureOrigin))
	signatureSha := hashedSecret.Sum(nil)

	signature := base64.StdEncoding.EncodeToString(signatureSha)

	authOrigin := fmt.Sprintf(`api_key="%s", algorithm="hmac-sha256", headers="host date request-line", signature="%s"`, e.Request.APIKey, signature)
	auth := base64.StdEncoding.EncodeToString([]byte(authOrigin))

	v := url.Values{}
	v.Set("authorization", auth)
	v.Set("date", date)
	v.Set("host", os.Getenv("XFYUN_TTS_BASE_HOST"))

	e.URL = os.Getenv("XFYUN_TTS_BASE_URL") + "?" + v.Encode()
}

func (e *SpeechSynthesizerImpl) BuildFile(vocabulary *models.LoquelaVocabulary) {
	e.Vocabulary = vocabulary
}

func (e *SpeechSynthesizerImpl) Synthesize(specifications string) (err error) {
	c, _, err := websocket.DefaultDialer.Dial(e.URL, nil)
	if err != nil {
		e.logger["EXTERNAL"].Printf("[ERROR][LOQUELA] Failed dialing websocket, %v", err)
	}
	defer c.Close()

	e.onOpen(c)
	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			e.onError(err)
			break
		}
		e.onMessage(c, messageType, message)
	}
	e.onClose(0, "")
	return
}

func (e *SpeechSynthesizerImpl) onMessage(ws *websocket.Conn, messageType int, message []byte) {
	if messageType != websocket.TextMessage {
		e.logger["EXTERNAL"].Printf("[ERROR][LOQUELA] Received non-text message: %d", messageType)
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(message, &response); err != nil {
		e.logger["EXTERNAL"].Printf("[ERROR][LOQUELA] Failed unmarshaling response, %v", err)
		return
	}

	code := int(response["code"].(float64))
	sid := response["sid"].(string)
	data := response["data"].(map[string]interface{})
	audioBase64 := data["audio"].(string)
	audio, err := base64.StdEncoding.DecodeString(audioBase64)
	if err != nil {
		e.logger["EXTERNAL"].Printf("[ERROR][LOQUELA] Failed decoding audio, %v", err)
		return
	}

	status := int(data["status"].(float64))
	if status == STATUS_LAST_FRAME {
		e.logger["EXTERNAL"].Printf("[INFO][LOQUELA] Closing Websocket.")
		ws.Close()
	}

	if code != 0 {
		errMsg := response["message"].(string)
		e.logger["EXTERNAL"].Printf("[ERROR][LOQUELA] Error Response from websocket, sid:%s call error:%s code is:%d", sid, errMsg, code)
	} else {
		file, err := os.OpenFile(e.Vocabulary.AudioPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			e.logger["EXTERNAL"].Printf("[ERROR][LOQUELA] Error opening file, %v", err)
			return
		}
		defer file.Close()
		file.Write(audio)
	}
}

func (e *SpeechSynthesizerImpl) onError(err error) {
	e.logger["EXTERNAL"].Printf("[ERROR][LOQUELA] Error while processing websocket request, Err: %v", err)
}

func (e *SpeechSynthesizerImpl) onClose(code int, text string) {
	e.logger["EXTERNAL"].Printf("[INFO][LOQUELA] Closed Websocket, Text: %s, Code: %d", text, code)
}

func (e *SpeechSynthesizerImpl) onOpen(ws *websocket.Conn) {
	d := map[string]interface{}{
		"common":   e.Request.CommonArgs,
		"business": e.Request.BusinessArgs,
		"data":     e.Request.Data,
	}
	dJson, _ := json.Marshal(d)
	e.logger["EXTERNAL"].Printf("[INFO][LOQUELA] Start sending data with websocket")
	ws.WriteMessage(websocket.TextMessage, dJson)

	if _, err := os.Stat(e.Vocabulary.AudioPath); err == nil {
		os.Remove(e.Vocabulary.AudioPath)
	}
}
