package gemini

import (
	"avalon/pkg/dto"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
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

func addWavHeader(pcmData []byte) []byte {
	sampleRate := 24000
	numChannels := 1
	bitsPerSample := 16

	dataLen := len(pcmData)
	header := make([]byte, 44)

	// Chunk ID
	copy(header[0:4], "RIFF")
	// Chunk Size
	binary.LittleEndian.PutUint32(header[4:8], uint32(36+dataLen))
	// Format
	copy(header[8:12], "WAVE")
	// Subchunk1 ID
	copy(header[12:16], "fmt ")
	// Subchunk1 Size
	binary.LittleEndian.PutUint32(header[16:20], 16)
	// Audio Format (1 = PCM)
	binary.LittleEndian.PutUint16(header[20:22], 1)
	// Num Channels
	binary.LittleEndian.PutUint16(header[22:24], uint16(numChannels))
	// Sample Rate
	binary.LittleEndian.PutUint32(header[24:28], uint32(sampleRate))
	// Byte Rate
	binary.LittleEndian.PutUint32(header[28:32], uint32(sampleRate*numChannels*bitsPerSample/8))
	// Block Align
	binary.LittleEndian.PutUint16(header[32:34], uint16(numChannels*bitsPerSample/8))
	// Bits Per Sample
	binary.LittleEndian.PutUint16(header[34:36], uint16(bitsPerSample))
	// Subchunk2 ID
	copy(header[36:40], "data")
	// Subchunk2 Size
	binary.LittleEndian.PutUint32(header[40:44], uint32(dataLen))

	return append(header, pcmData...)
}

type ttsRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
	GenerationConfig struct {
		ResponseModalities []string `json:"responseModalities"`
		Temperature        float64  `json:"temperature"` // Добавлено
		TopP               float64  `json:"topP"`        // Добавлено
		TopK               int      `json:"topK"`        // Добавлено
		SpeechConfig       struct {
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

	// 1. Приводим текст к нижнему регистру
	text = strings.ToLower(text)

	// 2. Убираем двоеточия (заменяем на тире для естественной паузы)
	text = strings.ReplaceAll(text, ":", " —")
	text = strings.ReplaceAll(text, "\"", " ")
	text = strings.ReplaceAll(text, "'", "")
	text = strings.ReplaceAll(text, "*", "")

	// 3. Убираем лишние переносы строк и двойные пробелы
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.Join(strings.Fields(text), " ")

	// 4. Гарантируем точку в самом конце, чтобы фраза не обрывалась
	if !strings.HasSuffix(text, ".") && !strings.HasSuffix(text, "!") && !strings.HasSuffix(text, "?") {
		text += "."
	}
	finalText := text
	if systemStyle != nil && *systemStyle != "" {
		finalText = *systemStyle + "\n\nЗачитай следующий текст полностью ничего не отбрасывая:\n" + text
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

	reqBody.GenerationConfig.ResponseModalities = []string{"AUDIO"}

	// Рекомендуемые настройки для стабильного TTS:
	reqBody.GenerationConfig.Temperature = 0.7 // Немного творчества для естественности голоса
	reqBody.GenerationConfig.TopP = 0.95       // Оставляем почти все вероятные токены
	reqBody.GenerationConfig.TopK = 40         // Стандартное значение для баланса

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

	return addWavHeader(audioBytes), nil
}
