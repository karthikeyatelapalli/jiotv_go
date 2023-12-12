package television

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/valyala/fasthttp"

	"github.com/rabilrbl/jiotv_go/v2/pkg/secureurl"
	"github.com/rabilrbl/jiotv_go/v2/pkg/utils"
)

const (
	// URL for fetching channels from JioTV API
	CHANNELS_API_URL = "https://jiotvapi.cdn.jio.com/apis/v3.0/getMobileChannelList/get/?langId=6&os=android&devicetype=phone&usertype=JIO&version=315&langId=6"
)

var (
	// DisableTSHandler is used to serve .ts files directly from JioTV Servers
	DisableTSHandler = os.Getenv("JIOTV_DISABLE_TS_HANDLER") == "true"
	SONY_CHANNELS = map[string]string{
		"sonyhd": "https://dai.google.com/linear/hls/event/dBdwOiGaQvy0TA1zOsjV6w/master.m3u8",
		"sonysabhd": "https://dai.google.com/linear/hls/event/CrTivkDESWqwvUj3zFEYEA/master.m3u8",
		"sonypal": "https://dai.google.com/linear/hls/event/dhPrGRwDRvuMQtmlzppzQQ/master.m3u8",
		"sonypixhd": "https://dai.google.com/linear/hls/event/x7rXWd2ERZ2tvyQWPmO1HA/master.m3u8",
		"sonymaxhd": "https://dai.google.com/linear/hls/event/UcjHNJmCQ1WRlGKlZm73QA/master.m3u8",
		"sonymax2": "https://dai.google.com/linear/hls/event/MdQ5Zy-PSraOccXu8jflCg/master.m3u8",
		"sonywah": "https://dai.google.com/linear/hls/event/gX5rCBf6Q7-D5AWY-sovzQ/master.m3u8",
		"sonyten1hd": "https://dai.google.com/linear/hls/event/wG75n5U8RrOKiFzaWObXbA/master.m3u8",
		"sonyten2hd": "https://dai.google.com/linear/hls/event/V9h-iyOxRiGp41ppQScDSQ/master.m3u8",
		"sonyten3hd": "https://dai.google.com/linear/hls/event/ltsCG7TBSCSDmyq0rQtvSA/master.m3u8",
		"sonyten4hd": "https://dai.google.com/linear/hls/event/smYybI_JToWaHzwoxSE9qA/master.m3u8",
		"sonyten5hd": "https://dai.google.com/linear/hls/event/Sle_TR8rQIuZHWzshEXYjQ/master.m3u8",
		"sonybbcearthhd": "https://dai.google.com/linear/hls/event/6bVWYIKGS0CIa-cOpZZJPQ/master.m3u8",
	}
	SONY_JIO_MAP = map[string]string{
		"sl291": "sonyhd",
		"sl154": "sonysabhd",
		"sl474": "sonypal",
		"sl762": "sonypixhd",
		"sl476": "sonymaxhd",
		"sl483": "sonymax2",
		"sl1393": "sonywah",
		"sl162": "sonyten1hd",
		"sl891": "sonyten2hd",
		"sl892": "sonyten3hd",
		"sl1772": "sonyten4hd",
		"sl155": "sonyten5hd",
		"sl852": "sonybbcearthhd",
	}
)

// New function creates a new Television instance with the provided credentials
func New(credentials *utils.JIOTV_CREDENTIALS) *Television {
	// Check if credentials are provided
	if credentials == nil {
		// If credentials are not provided, set them to empty strings
		credentials = &utils.JIOTV_CREDENTIALS{
			AccessToken: "",
			SSOToken:    "",
			CRM:         "",
			UniqueID:    "",
		}
	}
	headers := map[string]string{
		"Content-type": "application/x-www-form-urlencoded",
		"appkey":       "NzNiMDhlYzQyNjJm",
		"channel_id":   "",
		"crmid":        credentials.CRM,
		"userId":       credentials.CRM,
		"deviceId":     "e4286d7b481d69b8",
		"devicetype":   "phone",
		"isott":        "false",
		"languageId":   "6",
		"lbcookie":     "1",
		"os":           "android",
		"osVersion":    "13",
		"subscriberId": credentials.CRM,
		"uniqueId":     credentials.UniqueID,
		"User-Agent":   "okhttp/4.2.2",
		"usergroup":    "tvYR7NSNn7rymo3F",
		"versionCode":  "330",
	}

	// Create a fasthttp.Client
	client := utils.GetRequestClient()

	// Return a new Television instance
	return &Television{
		AccessToken: credentials.AccessToken,
		SsoToken:    credentials.SSOToken,
		Crm:         credentials.CRM,
		UniqueID:    credentials.UniqueID,
		Headers:     headers,
		Client:      client,
	}
}

// Live method generates m3u8 link from JioTV API with the provided channel ID
func (tv *Television) Live(channelID string) (*LiveURLOutput, error) {
	// If channelID starts with sl, then it is a Sony Channel
	if channelID[:2] == "sl" {
		return GetSLChannel(channelID)
	}

	formData := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(formData)

	formData.Add("channel_id", channelID)
	formData.Add("stream_type", "Seek")
	formData.Add("begin", utils.GenerateCurrentTime())
	formData.Add("srno", utils.GenerateDate())

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	// Copy headers from the Television headers map to the request
	for key, value := range tv.Headers {
		req.Header.Set(key, value)
	}

	var url string
	if tv.AccessToken != "" {
		url = "https://jiotvapi.media.jio.com/playback/apis/v1/geturl?langId=6"
		req.Header.Set("accesstoken", tv.AccessToken)
	} else {
		req.Header.Set("osVersion", "8.1.0")
		req.Header.Set("ssotoken", tv.SsoToken)
		req.Header.Set("versionCode", "277")
		url = "https://tv.media.jio.com/apis/v2.2/getchannelurl/getchannelurl"
		req.Header.SetUserAgent("plaYtv/7.0.5 (Linux;Android 8.1.0) ExoPlayerLib/2.11.7")
	}
	req.SetRequestURI(url)
	req.Header.SetMethod("POST")

	// Encode the form data and set it as the request body
	req.SetBody(formData.QueryString())

	req.Header.Set("channel_id", channelID)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Perform the HTTP POST request
	if err := tv.Client.Do(req, resp); err != nil {
		utils.Log.Panic(err)
		return nil, err
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		// Store the response body as a string
		response := string(resp.Body())

		// Log headers and request data
		utils.Log.Println("Request headers:", req.Header.String())
		utils.Log.Println("Request data:", formData.String())
		utils.Log.Panicln("Response: ", response)

		return nil, fmt.Errorf("Request failed with status code: %d\nresponse: %s", resp.StatusCode(), response)
	}

	var result LiveURLOutput
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		utils.Log.Panic(err)
		return nil, err
	}

	return &result, nil
}

// Render method does HTTP GET request to the provided URL and return the response body
func (tv *Television) Render(url string) ([]byte, int) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)
	req.Header.SetMethod("GET")

	// Copy headers from the Television headers map to the request
	for key, value := range tv.Headers {
		req.Header.Set(key, value)
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Perform the HTTP GET request
	if err := tv.Client.Do(req, resp); err != nil {
		utils.Log.Panic(err)
	}

	buf := resp.Body()

	return buf, resp.StatusCode()
}

// Channels fetch channels from JioTV API
func Channels() ChannelsResponse {

	// Create a fasthttp.Client
	client := utils.GetRequestClient()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(CHANNELS_API_URL)

	req.Header.SetMethod("GET")
	req.Header.Add("User-Agent", "okhttp/4.2.2")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("devicetype", "phone")
	req.Header.Add("os", "android")
	req.Header.Add("appkey", "NzNiMDhlYzQyNjJm")
	req.Header.Add("lbcookie", "1")
	req.Header.Add("usertype", "JIO")

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Perform the HTTP GET request
	if err := client.Do(req, resp); err != nil {
		utils.Log.Panic(err)
	}

	var apiResponse ChannelsResponse

	// Check the response status code
	if resp.StatusCode() != fasthttp.StatusOK {
		utils.Log.Panicf("Request failed with status code: %d", resp.StatusCode())
	}

	resp_body, err := resp.BodyGunzip()
	if err != nil {
		utils.Log.Panic(err)
	}

	// Parse the JSON response
	if err := json.Unmarshal(resp_body, &apiResponse); err != nil {
		utils.Log.Panic(err)
	}

	return apiResponse
}

// FilterChannels Function is used to filter channels by language and category
func FilterChannels(channels []Channel, language, category int) []Channel {
	var filteredChannels []Channel
	for _, channel := range channels {
		// if both language and category is set, then use and operator
		if language != 0 && category != 0 {
			if channel.Language == language && channel.Category == category {
				filteredChannels = append(filteredChannels, channel)
			}
		} else if language != 0 {
			if channel.Language == language {
				filteredChannels = append(filteredChannels, channel)
			}
		} else if category != 0 {
			if channel.Category == category {
				filteredChannels = append(filteredChannels, channel)
			}
		} else {
			filteredChannels = append(filteredChannels, channel)
		}
	}
	return filteredChannels
}

func ReplaceM3U8(baseUrl, match []byte, params, channel_id string) []byte {
	coded_url, err := secureurl.EncryptURL(string(baseUrl) + string(match) + "?" + params)
	if err != nil {
		utils.Log.Println(err)
		return nil
	}
	return []byte("/render.m3u8?auth=" + coded_url + "&channel_key_id=" + channel_id)
}

func ReplaceTS(baseUrl, match []byte, params string) []byte {
	if DisableTSHandler {
		return []byte(string(baseUrl) + string(match) + "?" + params)
	}
	coded_url, err := secureurl.EncryptURL(string(baseUrl) + string(match) + "?" + params)
	if err != nil {
		utils.Log.Println(err)
		return nil
	}
	return []byte("/render.ts?auth=" + coded_url)
}

func ReplaceAAC(baseUrl, match []byte, params string) []byte {
	if DisableTSHandler {
		return []byte(string(baseUrl) + string(match) + "?" + params)
	}
	coded_url, err := secureurl.EncryptURL(string(baseUrl) + string(match) + "?" + params)
	if err != nil {
		utils.Log.Println(err)
		return nil
	}
	return []byte("/render.ts?auth=" + coded_url)
}

func ReplaceKey(match []byte, params, channel_id string) []byte {
	coded_url, err := secureurl.EncryptURL(string(match) + "?" + params)
	if err != nil {
		utils.Log.Println(err)
		return nil
	}
	return []byte("/render.key?auth=" + coded_url + "&channel_key_id=" + channel_id)
}

func GetSLChannel(channelID string) (*LiveURLOutput, error) {
	// Check if the channel is available in the SONY_CHANNELS map
	if val, ok := SONY_JIO_MAP[channelID]; ok {
		// If the channel is available in the SONY_CHANNELS map, then return the link
		fmt.Println(val)
		fmt.Println(SONY_CHANNELS[val])
		result := new(LiveURLOutput)

		channel_url := SONY_CHANNELS[val]		

		// Make a get request to the channel url and store location header in actual_url
		req := fasthttp.AcquireRequest()
		defer fasthttp.ReleaseRequest(req)

		req.SetRequestURI(channel_url)
		req.Header.SetMethod("GET")

		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseResponse(resp)

		// Perform the HTTP GET request
		if err := utils.GetRequestClient().Do(req, resp); err != nil {
			utils.Log.Panic(err)
		}

		if resp.StatusCode() != fasthttp.StatusOK {
			utils.Log.Panicf("Request failed with status code: %d", resp.StatusCode())
			utils.Log.Panicln("Response: ", string(resp.Body()))
		}

		// Store the location header in actual_url
		actual_url := string(resp.Header.Peek("Location"))

		result.Result = actual_url
		result.Bitrates.Auto = actual_url
		return result, nil
	} else {
		// If the channel is not available in the SONY_CHANNELS map, then return an error
		return nil, fmt.Errorf("Channel not found")
	}
}
