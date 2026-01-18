package gemini

import (
	"avalon/pkg/dto"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type TtsGeminiAgent struct {
	apiKey     string
	httpClient *http.Client
}

// New создаёт Gemini-клиент
func NewTtsAgent(apiKey string) dto.TtsAgent {
	return &TtsGeminiAgent{
		apiKey:     apiKey,
		httpClient: http.DefaultClient,
	}
}

type ttsRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
	GenerationConfig struct {
		SpeechConfig struct {
			VoiceConfig struct {
				PrebuiltVoiceConfig struct {
					VoiceName string `json:"voiceName"`
				} `json:"prebuiltVoiceConfig"`
			} `json:"voiceConfig"`
		} `json:"speechConfig"`
	} `json:"generationConfig"`
}

type ttsResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				InlineData struct {
					MimeType string `json:"mimeType"`
					Data     string `json:"data"`
				} `json:"inlineData"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// Send отправляет текст в Gemini TTS и возвращает аудиобайты
func (c *TtsGeminiAgent) Send(
	model string,
	voiceName string,
	text string,
	systemStyle *string,
) ([]byte, error) {

	if c.apiKey == "" {
		return nil, errors.New("apiKey is empty")
	}
	if model == "" {
		return nil, errors.New("model is empty")
	}
	if voiceName == "" {
		return nil, errors.New("voiceName is empty")
	}
	if text == "" {
		return nil, errors.New("text is empty")
	}

	// Формируем итоговый текст (systemStyle не озвучивается)
	finalText := text
	if systemStyle != nil && *systemStyle != "" {
		finalText = *systemStyle + "\n\nТекст:\n" + text
	}

	// Запрос
	var reqBody ttsRequest
	reqBody.Contents = append(reqBody.Contents, struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	}{
		Parts: []struct {
			Text string `json:"text"`
		}{
			{Text: finalText},
		},
	})

	reqBody.GenerationConfig.
		SpeechConfig.
		VoiceConfig.
		PrebuiltVoiceConfig.
		VoiceName = voiceName

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/" +
		model + ":generateContent?key=" + c.apiKey

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(raw))
	}

	var ttsResp ttsResponse
	if err := json.Unmarshal(raw, &ttsResp); err != nil {
		return nil, err
	}

	if len(ttsResp.Candidates) == 0 ||
		len(ttsResp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("empty TTS response")
	}

	audioBase64 := ttsResp.
		Candidates[0].
		Content.
		Parts[0].
		InlineData.
		Data

	audioBytes, err := base64.StdEncoding.DecodeString(audioBase64)
	if err != nil {
		return nil, err
	}

	return audioBytes, nil
}
