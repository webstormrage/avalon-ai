package elevenLabs

import (
	"avalon/pkg/dto"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

type TtsElevenLabsAgent struct {
	apiKey     string
	httpClient *http.Client
}

// New создаёт ElevenLabs-клиент
func NewTtsAgent(apiKey string) dto.TtsAgent {
	return &TtsElevenLabsAgent{
		apiKey:     apiKey,
		httpClient: http.DefaultClient,
	}
}

type ttsRequest struct {
	Text          string `json:"text"`
	ModelID       string `json:"model_id"`
	VoiceSettings struct {
		Stability       float64 `json:"stability"`
		SimilarityBoost float64 `json:"similarity_boost"`
	} `json:"voice_settings"`
}

// Send отправляет текст в ElevenLabs TTS и возвращает WAV/MP3
func (c *TtsElevenLabsAgent) Send(
	persona dto.PlayerV2,
	text string,
	systemStyle *string,
) ([]byte, error) {

	if c.apiKey == "" {
		return nil, errors.New("apiKey is empty")
	}
	if persona.Voice == "" {
		return nil, errors.New("voiceId is empty")
	}
	if persona.TtsModel == "" {
		return nil, errors.New("model is empty")
	}
	if text == "" {
		return nil, errors.New("text is empty")
	}

	// ---- 1. Подготовка текста (как у тебя)
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, ":", " —")
	text = strings.ReplaceAll(text, "\"", " ")
	text = strings.ReplaceAll(text, "'", "")
	text = strings.ReplaceAll(text, "*", "")
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.Join(strings.Fields(text), " ")

	if !strings.HasSuffix(text, ".") &&
		!strings.HasSuffix(text, "!") &&
		!strings.HasSuffix(text, "?") {
		text += "."
	}

	finalText := text
	/*if systemStyle != nil && *systemStyle != "" {
		finalText = *systemStyle + "\n\n" + text
	}*/

	// ---- 2. Формируем запрос
	var reqBody ttsRequest
	reqBody.Text = finalText
	reqBody.ModelID = persona.TtsModel // например: eleven_multilingual_v2
	reqBody.VoiceSettings.Stability = float64(1.0 - persona.VoiceTemperature)
	reqBody.VoiceSettings.SimilarityBoost = 0.75 // можно вынести в persona

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := "https://api.elevenlabs.io/v1/text-to-speech/" + persona.Voice

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("Accept", "audio/mpeg") // или audio/wav

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	audioBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(audioBytes))
	}

	// ElevenLabs уже возвращает готовое аудио
	return audioBytes, nil
}
