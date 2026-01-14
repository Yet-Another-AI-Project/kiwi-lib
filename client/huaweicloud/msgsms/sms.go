package msgsms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/Yet-Another-AI-Project/kiwi-lib/client/huaweicloud/core"
	"github.com/futurxlab/golanggraph/xerror"
)

type SendSMSResponse struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Result      string `json:"result"`
}

func (h *HuaweiCloudSMS) SendSMS(receiver string, templateParams map[string]string) (*SendSMSResponse, error) {
	appInfo := core.Signer{
		Key:    h.key,
		Secret: h.secret,
	}

	url := fmt.Sprintf("%s/sms/batchSendSms/v1", h.endpoint)

	tp, err := json.Marshal(templateParams)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	body := h.buildRequestBody(receiver, string(tp))
	resp, err := h.post(url, []byte(body), appInfo)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	var response SendSMSResponse
	err = json.Unmarshal(resp, &response)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return &response, nil
}

func (h *HuaweiCloudSMS) buildRequestBody(receiver, templateParas string) string {
	param := "from=" + url.QueryEscape(h.sender) + "&to=" + url.QueryEscape(receiver) + "&templateId=" + url.QueryEscape(h.templateId)
	if templateParas != "" {
		param += "&templateParas=" + url.QueryEscape(templateParas)
	}
	if h.signature != "" {
		param += "&signature=" + url.QueryEscape(h.signature)
	}
	return param
}

func (h *HuaweiCloudSMS) post(url string, param []byte, appInfo core.Signer) ([]byte, error) {
	if param == nil || appInfo == (core.Signer{}) {
		return nil, fmt.Errorf("param or appInfo is nil")
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(param))
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	// Add the content format to the request. The header is fixed.
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// Sign the request using the HMAC algorithm and add the signing result to the Authorization header.
	appInfo.Sign(req)

	// Send an SMS request.
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	// Response to obtaining the SMS message
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, xerror.Wrap(err)
	}
	return body, nil
}
