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

	// You can send signaling messages back to clients using conn.WriteMessage
}

func handleOffer(conn *websocket.Conn, offerSDP string, peerConnection *webrtc.PeerConnection) error {

	// Parse the received offer SDP
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offerSDP,
	}

	// Set the offer as the remote description
	err := peerConnection.SetRemoteDescription(offer)
	if err != nil {
		return errors.Join(errors.New("setRemote"), err)
	}

	// Create audio and video tracks from user media
	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "pion")
	if err != nil {
		return errors.Join(errors.New("setTrackA"), err)
	}

	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}, "video", "pion")
	if err != nil {
		return errors.Join(errors.New("setTrackV"), err)
	}

	// Add tracks to the peer connection
	_, err = peerConnection.AddTrack(audioTrack)
	if err != nil {
		return errors.Join(errors.New("AddtrackA"), err)
	}

	_, err = peerConnection.AddTrack(videoTrack)
	if err != nil {
		return errors.Join(errors.New("AddtrackV"), err)
	}

	// Create an answer SDP
	answerSDP, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return errors.Join(errors.New("CreateAnswer"), err)
	}

	// Set the answer as the local description
	err = peerConnection.SetLocalDescription(answerSDP)
	if err != nil {
		return errors.Join(errors.New("setAnswer"), err)
	}

	// Convert the answer SDP to JSON
	answerSDPBytes, err := json.Marshal(answerSDP)
	if err != nil {
		return errors.Join(errors.New("json"), err)
	}

	// Send the answer SDP back to the client
	return conn.WriteMessage(websocket.TextMessage, answerSDPBytes)
}

func handleAnswer(conn *websocket.Conn, answerSDP string, peerConnection *webrtc.PeerConnection) error {

	answer := webrtc.SessionDescription{}
	err := json.Unmarshal([]byte(answerSDP), &answer)
	if err != nil {
		return errors.Join(errors.New("json"), err)
	}

	err = peerConnection.SetRemoteDescription(answer)
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
