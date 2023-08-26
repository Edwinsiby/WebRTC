package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

var signalingMsg struct {
	Type string `json:"type"`
	Data string `json:"data"`
}
var StoredOffer string

func GetStoredOffer() string {
	return StoredOffer
}

type WebRTCHandler struct{}

func NewWebRTCHandler() *WebRTCHandler {
	return &WebRTCHandler{}
}

func (h *WebRTCHandler) SetupRoutes(router *gin.RouterGroup) {
	router.GET("/ws", h.handleWebSocket)
}

func (h *WebRTCHandler) handleWebSocket(c *gin.Context) {

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// Allow all origins for WebSocket connections
			return true
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// Handle error
		fmt.Println("error upgrade", err)
		return
	}
	defer conn.Close()

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("error readmsg", err)
			break
		}

		if err := json.Unmarshal(message, &signalingMsg); err != nil {
			fmt.Println("error json:", err)
			continue
		}

		switch signalingMsg.Type {
		case "offer":
			fmt.Println("case offer")
			err := handleOffer(conn, signalingMsg.Data, peerConnection)
			if err != nil {
				fmt.Println("error case1", err.Error())
				break
			}
		case "answer":
			fmt.Println("case answer")
			err := handleAnswer(conn, signalingMsg.Data, peerConnection)
			if err != nil {
				fmt.Println("error case2", err)
				break
			}
		case "candidate":
			fmt.Println("case candidate")
			err := handleICECandidate(conn, signalingMsg.Data, peerConnection)
			if err != nil {
				fmt.Println("error case3", err)
				break
			}
		default:
			fmt.Println("error", "unknown type")
		}
	}

}
func handleOffer(conn *websocket.Conn, offerSDPstring string, peerConnection *webrtc.PeerConnection) error {
	offerSDP := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offerSDPstring,
	}

	StoredOffer = offerSDPstring

	err := peerConnection.SetRemoteDescription(offerSDP)
	if err != nil {
		return errors.Join(errors.New("setRemote"), err)
	}

	offerSDPBytes, err := json.Marshal(offerSDP)
	if err != nil {
		return errors.Join(errors.New("json"), err)
	}

	return conn.WriteMessage(websocket.TextMessage, offerSDPBytes)
}

func handleAnswer(conn *websocket.Conn, answerSDP string, peerConnection *webrtc.PeerConnection) error {

	answer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  answerSDP,
	}

	err := peerConnection.SetRemoteDescription(answer)
	if err != nil {
		return errors.Join(errors.New("SetRemoteDsc"), err)
	}

	return nil
}

func handleICECandidate(conn *websocket.Conn, candidate string, peerConnection *webrtc.PeerConnection) error {

	iceCandidate := webrtc.ICECandidateInit{
		Candidate: candidate,
	}

	// Add the ICE candidate to the peer connection
	err := peerConnection.AddICECandidate(iceCandidate)
	if err != nil {
		return errors.Join(errors.New("AddICE"), err)
	}

	return nil
}

func getUserMedia() (*webrtc.TrackLocalStaticSample, *webrtc.TrackLocalStaticSample, error) {
	// Create a new audio track
	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio_track", "audio_stream")
	if err != nil {
		return nil, nil, err
	}

	// Create a new video track
	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}, "video_track", "video_stream")
	if err != nil {
		return nil, nil, err
	}

	// Return the audio track and video track
	return audioTrack, videoTrack, nil
}
